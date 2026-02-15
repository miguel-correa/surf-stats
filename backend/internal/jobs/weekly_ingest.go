package jobs

import (
	"database/sql"
	"fmt"
	"log"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/maps"
	"surfstats/internal/scrapers/players"
	"sync"
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
				mapID, comps, err := playerScraper.FetchMapCompletionsFromPlayerRecord(seedSteamID, job.Name)
				resultsCh <- completionResult{
					MapName: job.Name,
					MapID:   mapID,
					Comps:   comps,
					Err:     err,
				}
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
	for result := range resultsCh {
		processed++

		if result.Err != nil {
			return fmt.Errorf("error fetching completions for map %q: %w", result.MapName, result.Err)
		}

		if err := db.UpdateMapCompletionsByMapIDTx(tx, result.MapID, result.Comps); err != nil {
			return fmt.Errorf("update completions map_id=%d: %w", result.MapID, err)
		}

		if processed%25 == 0 || processed == len(scrapedMaps) {
			log.Printf("weekly-ingest: %d/%d processed", processed, len(scrapedMaps))
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	log.Printf("weekly-ingest: done")
	return nil
}
