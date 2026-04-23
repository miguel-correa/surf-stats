package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
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
	return runPlayerRecordsIngestionWithWorkers(database, steamID, players.NewKSFScraper(), 16)
}

func runPlayerRecordsIngestion(database *sql.DB, steamID string, scraper mainRecordScraper) error {
	return runPlayerRecordsIngestionWithWorkers(database, steamID, scraper, 16)
}

func runPlayerRecordsIngestionWithWorkers(database *sql.DB, steamID string, scraper mainRecordScraper, workerCount int) error {
	maps, err := db.ListMapsForPlayerRecordIngestion(database)
	if err != nil {
		return err
	}
	totalMaps := len(maps)
	log.Printf("player-records-ingest: start steam_id=%s maps=%d", steamID, totalMaps)

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
	var player *models.Player
	processed := 0
	for result := range resultsCh {
		processed++
		if result.Err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, result.Err))
			log.Printf("player-records-ingest: skipping map=%s after error: %v", result.MapName, result.Err)
		} else if result.Record != nil {
			if player == nil {
				playerID := result.Record.PlayerID
				player = &models.Player{
					SteamID:  result.Record.SteamID,
					PlayerID: &playerID,
					Name:     result.Record.PlayerName,
				}
			}

			inserted, err := db.SavePlayerMapRecordIfImproved(database, *result.Record)
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", result.MapName, err))
			} else if inserted {
				log.Printf("player-records-ingest: stored PB steam_id=%s map=%s time_ms=%d", steamID, result.MapName, result.Record.SurfTimeMS)
			}
		}

		if processed%25 == 0 || processed == totalMaps {
			log.Printf("player-records-ingest: %d/%d processed", processed, totalMaps)
		}
	}

	if player != nil {
		if err := db.UpsertPlayer(database, *player); err != nil {
			failures = append(failures, fmt.Sprintf("player %s: %v", player.SteamID, err))
			log.Printf("player-records-ingest: failed to upsert player steam_id=%s: %v", player.SteamID, err)
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

	if err == context.DeadlineExceeded {
		return true
	}

	if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		return true
	}

	r, ok := err.(retryable)
	return ok && r.Retryable()
}
