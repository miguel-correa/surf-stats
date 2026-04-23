package jobs

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/maps"
	"surfstats/internal/scrapers/players"
	"sync"
	"time"
)

type completionJob struct {
	Name string
}

type completionResult struct {
	MapName string
	MapID   int
	Comps   int
	Err     error
}

func RunWeeklyIngestion(database *sql.DB, seedSteamID string) error {
	log.Printf("weekly-ingest: start")
	mapScraper := maps.NewKSFScraper("https://ksf.surf/maps")
	scrapedMaps, err := mapScraper.FetchMaps("")
	if err != nil {
		return fmt.Errorf("failed to fetch maps: %v", err)
	}
	log.Printf("weekly-ingest: fetched maps=%d", len(scrapedMaps))

	err = db.UpsertMaps(database, scrapedMaps)
	if err != nil {
		return fmt.Errorf("error upserting maps: %v", err)
	}

	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	const workerCount = 8
	jobsCh := make(chan completionJob)
	resultsCh := make(chan completionResult, workerCount)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			playerScraper := players.NewKSFScraper()
			for job := range jobsCh {
				mapID, comps, err := fetchCompletionsWithRetry(playerScraper, seedSteamID, job.Name)
				resultsCh <- completionResult{
					MapName: job.Name,
					MapID:   mapID,
					Comps:   comps,
					Err:     err,
				}
				time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(150))*time.Millisecond)
			}
		}()
	}

	go func() {
		for _, m := range scrapedMaps {
			jobsCh <- completionJob{Name: m.Name}
		}
		close(jobsCh)
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	processed := 0
	var failures []string
	for result := range resultsCh {
		processed++

		if result.Err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, result.Err))
			log.Printf("weekly-ingest: skipping map=%s: %v", result.MapName, result.Err)
		} else if err := db.UpdateMapCompletionsByMapIDTx(tx, result.MapID, result.Comps); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, err))
			log.Printf("weekly-ingest: failed to update map=%s: %v", result.MapName, err)
		}

		if processed%25 == 0 || processed == len(scrapedMaps) {
			log.Printf("weekly-ingest: %d/%d processed", processed, len(scrapedMaps))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}

	if len(failures) > 0 {
		log.Printf("weekly-ingest: completed with %d skipped map(s); first failure: %s", len(failures), failures[0])
	}

	log.Printf("weekly-ingest: done")
	return nil
}

func fetchCompletionsWithRetry(scraper *players.KSFScraper, steamID, mapName string) (int, int, error) {
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		mapID, comps, err := scraper.FetchMapCompletionsFromPlayerRecord(steamID, mapName)
		if err == nil {
			return mapID, comps, nil
		}
		lastErr = err
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt*500) * time.Millisecond)
		}
	}
	return -1, -1, lastErr
}