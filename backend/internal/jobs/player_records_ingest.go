package jobs

import (
	"database/sql"
	"fmt"
	"log"

	"surfstats/internal/db"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/players"
)

type mainRecordScraper interface {
	FetchMainMapRecord(steamID string, mapName string) (*models.PlayerMapRecord, error)
}

func RunPlayerRecordsIngestion(database *sql.DB, steamID string) error {
	return runPlayerRecordsIngestion(database, steamID, players.NewKSFScraper())
}

func runPlayerRecordsIngestion(database *sql.DB, steamID string, scraper mainRecordScraper) error {
	maps, err := db.ListMapsForPlayerRecordIngestion(database)
	if err != nil {
		return err
	}

	var failures []string
	for _, m := range maps {
		record, err := scraper.FetchMainMapRecord(steamID, m.Name)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		if record == nil {
			continue
		}

		inserted, err := db.SavePlayerMapRecordIfImproved(database, *record)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", m.Name, err))
			continue
		}
		if inserted {
			log.Printf("player-records-ingest: stored PB steam_id=%s map=%s time_ms=%d", steamID, m.Name, record.SurfTimeMS)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("player-records-ingest: %d map(s) failed; first failure: %s", len(failures), failures[0])
	}

	return nil
}
