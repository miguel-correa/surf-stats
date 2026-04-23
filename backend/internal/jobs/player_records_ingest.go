package jobs

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/players"
)

type mainRecordScraper interface {
	FetchMainMapRecord(steamID string, mapName string) (*models.PlayerMapRecord, error)
}

type playerRecordJob struct {
	KSFMapID int
	MapName  string
}

type playerRecordResult struct {
	MapName string
	Record  *models.PlayerMapRecord
	Err     error
}

func RunPlayerRecordsIngestion(database *sql.DB, steamID string) error {
	return runPlayerRecordsIngestionWithWorkers(database, steamID, players.NewKSFScraper(), 4)
}

func runPlayerRecordsIngestion(database *sql.DB, steamID string, scraper mainRecordScraper) error {
	return runPlayerRecordsIngestionWithWorkers(database, steamID, scraper, 4)
}

func runPlayerRecordsIngestionWithWorkers(database *sql.DB, steamID string, scraper mainRecordScraper, workerCount int) error {
	maps, err := db.ListMapsForPlayerRecordIngestion(database)
	if err != nil {
		return err
	}

	if workerCount < 1 {
		workerCount = 1
	}

	jobsCh := make(chan playerRecordJob)
	resultsCh := make(chan playerRecordResult, workerCount)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsCh {
				record, err := fetchMainMapRecordWithRetry(scraper, steamID, job.MapName)
				resultsCh <- playerRecordResult{
					MapName: job.MapName,
					Record:  record,
					Err:     err,
				}
				time.Sleep(100*time.Millisecond + time.Duration(rand.Intn(150))*time.Millisecond)
			}
		}()
	}

	go func() {
		for _, m := range maps {
			jobsCh <- playerRecordJob{
				KSFMapID: m.KSFMapID,
				MapName:  m.Name,
			}
		}
		close(jobsCh)
	}()

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var failures []string
	for result := range resultsCh {
		if result.Err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, result.Err))
			log.Printf("player-records-ingest: skipping map=%s after error: %v", result.MapName, result.Err)
			continue
		}
		if result.Record == nil {
			continue
		}

		playerID := result.Record.PlayerID
		if err := db.UpsertPlayer(database, models.Player{
			SteamID:  result.Record.SteamID,
			PlayerID: &playerID,
			Name:     result.Record.PlayerName,
		}); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, err))
			log.Printf("player-records-ingest: skipping map=%s after player upsert error: %v", result.MapName, err)
			continue
		}

		inserted, err := db.SavePlayerMapRecordIfImproved(database, *result.Record)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, err))
			continue
		}
		if inserted {
			log.Printf("player-records-ingest: stored PB steam_id=%s map=%s time_ms=%d", steamID, result.MapName, result.Record.SurfTimeMS)
		}
	}

	if len(failures) > 0 {
		log.Printf("player-records-ingest: completed with %d skipped map(s); first failure: %s", len(failures), failures[0])
	}

	return nil
}

func fetchMainMapRecordWithRetry(scraper mainRecordScraper, steamID string, mapName string) (*models.PlayerMapRecord, error) {
	const maxAttempts = 3

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		record, err := scraper.FetchMainMapRecord(steamID, mapName)
		if err == nil {
			return record, nil
		}

		lastErr = err
		if !isRetryablePlayerRecordError(err) || attempt == maxAttempts {
			break
		}

		time.Sleep(time.Duration(attempt*250) * time.Millisecond)
	}

	return nil, lastErr
}

func isRetryablePlayerRecordError(err error) bool {
	type retryable interface {
		Retryable() bool
	}

	if err == nil {
		return false
	}

	r, ok := err.(retryable)
	return ok && r.Retryable()
}
