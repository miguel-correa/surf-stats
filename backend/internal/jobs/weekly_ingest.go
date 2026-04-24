package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
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

func RunWeeklyIngestion(database *sql.DB, seedSteamID string) ([]string, error) {
	log.Printf("weekly-ingest: start")
	mapScraper := maps.NewKSFScraper("https://ksf.surf/maps")
	scrapedMaps, err := mapScraper.FetchMaps("")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch maps: %v", err)
	}
	log.Printf("weekly-ingest: fetched maps=%d", len(scrapedMaps))

	if err := db.UpsertMaps(database, scrapedMaps); err != nil {
		return nil, fmt.Errorf("error upserting maps: %v", err)
	}

	names := make([]string, len(scrapedMaps))
	for i, m := range scrapedMaps {
		names[i] = m.Name
	}
	return runCompletions(database, seedSteamID, names)
}

func RunCompletionsForMaps(database *sql.DB, seedSteamID string, mapNames []string) ([]string, error) {
	log.Printf("weekly-ingest: retry start maps=%d", len(mapNames))
	return runCompletions(database, seedSteamID, mapNames)
}

func runCompletions(database *sql.DB, seedSteamID string, mapNames []string) ([]string, error) {
	tx, err := database.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	const workerCount = 4
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
		for _, name := range mapNames {
			jobsCh <- completionJob{Name: name}
		}
		close(jobsCh)
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	total := len(mapNames)
	processed := 0
	var failedMaps []string
	for result := range resultsCh {
		processed++

		if result.Err != nil {
			failedMaps = append(failedMaps, result.MapName)
			log.Printf("weekly-ingest: skipping map=%s: %v", result.MapName, result.Err)
		} else if err := db.UpdateMapCompletionsByMapIDTx(tx, result.MapID, result.Comps); err != nil {
			failedMaps = append(failedMaps, result.MapName)
			log.Printf("weekly-ingest: failed to update map=%s: %v", result.MapName, err)
		}

		if processed%25 == 0 || processed == total {
			log.Printf("weekly-ingest: %d/%d processed", processed, total)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit tx: %w", err)
	}

	if len(failedMaps) > 0 {
		log.Printf("weekly-ingest: completed with %d failed map(s)", len(failedMaps))
	}

	log.Printf("weekly-ingest: done")
	return failedMaps, nil
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
		if !isRetryableError(err) || attempt == maxAttempts {
			break
		}
		backoff := time.Duration(1<<uint(attempt-1)) * time.Second
		time.Sleep(backoff)
	}
	return -1, -1, lastErr
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if err == context.DeadlineExceeded {
		return true
	}
	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}
	type retryable interface {
		Retryable() bool
	}
	if r, ok := err.(retryable); ok {
		return r.Retryable()
	}
	return false
}