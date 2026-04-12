package jobs

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/models"
)

func TestRunPlayerRecordsIngestionStoresOnlyNewOrImprovedMainRecords(t *testing.T) {
	database := openJobTestDatabase(t)

	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES
			(1, 'surf_alpha', 1),
			(2, 'surf_beta', 2),
			(3, 'surf_gamma', 3)
	`); err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}

	initialDate := time.Date(2026, time.April, 1, 12, 0, 0, 0, time.UTC)
	inserted, err := db.SavePlayerMapRecordIfImproved(database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 60000,
		Rank:       100,
		TotalRanks: 1000,
		DateSet:    &initialDate,
	})
	if err != nil {
		t.Fatalf("initial save failed: %v", err)
	}
	if !inserted {
		t.Fatal("initial save inserted = false, want true")
	}

	scraper := stubMainRecordScraper{
		recordsByMapName: map[string]*models.PlayerMapRecord{
			"surf_alpha": {
				SteamID:    "STEAM_0:1:75949009",
				PlayerID:   949217,
				KSFMapID:   1,
				SurfTimeMS: 45000,
				Rank:       50,
				TotalRanks: 500,
			},
			"surf_beta": {
				SteamID:    "STEAM_0:1:75949009",
				PlayerID:   949217,
				KSFMapID:   2,
				SurfTimeMS: 59000,
				Rank:       90,
				TotalRanks: 1000,
			},
			"surf_gamma": nil,
		},
	}

	if err := runPlayerRecordsIngestion(database, "STEAM_0:1:75949009", scraper); err != nil {
		t.Fatalf("runPlayerRecordsIngestion returned error: %v", err)
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM player_map_records`).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 3 {
		t.Fatalf("record count = %d, want %d", count, 3)
	}
}

type stubMainRecordScraper struct {
	recordsByMapName map[string]*models.PlayerMapRecord
}

func (s stubMainRecordScraper) FetchMainMapRecord(_ string, mapName string) (*models.PlayerMapRecord, error) {
	record, ok := s.recordsByMapName[mapName]
	if !ok {
		return nil, nil
	}
	return record, nil
}

func openJobTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	database := db.Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	for _, path := range []string{
		jobMigrationPath(t, "../../migrations/001_create_maps.sql"),
		jobMigrationPath(t, "../../migrations/002_create_player_map_records.sql"),
	} {
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%q) failed: %v", path, err)
		}
		if _, err := database.Exec(string(sqlBytes)); err != nil {
			t.Fatalf("Exec migration %q failed: %v", path, err)
		}
	}

	return database
}

func jobMigrationPath(t *testing.T, relative string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), relative))
}
