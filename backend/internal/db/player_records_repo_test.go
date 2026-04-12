package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"surfstats/internal/models"
)

func TestSavePlayerMapRecordIfImproved(t *testing.T) {
	database := openTestDatabase(t)

	initialDate := time.Date(2026, time.April, 10, 19, 13, 42, 0, time.UTC)
	initial := models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   285,
		SurfTimeMS: 105186,
		Rank:       142,
		TotalRanks: 6614,
		DateSet:    &initialDate,
	}

	inserted, err := SavePlayerMapRecordIfImproved(database, initial)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved returned error: %v", err)
	}
	if !inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = false, want true")
	}

	sameTime := initial
	sameTime.Rank = 100

	inserted, err = SavePlayerMapRecordIfImproved(database, sameTime)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved same time returned error: %v", err)
	}
	if inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = true for same time, want false")
	}

	worseTime := initial
	worseTime.SurfTimeMS = 120000

	inserted, err = SavePlayerMapRecordIfImproved(database, worseTime)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved worse time returned error: %v", err)
	}
	if inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = true for worse time, want false")
	}

	improvedDate := time.Date(2026, time.April, 12, 12, 0, 0, 0, time.UTC)
	improved := initial
	improved.SurfTimeMS = 104500
	improved.Rank = 130
	improved.DateSet = &improvedDate

	inserted, err = SavePlayerMapRecordIfImproved(database, improved)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved improved time returned error: %v", err)
	}
	if !inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = false for improved time, want true")
	}

	latest, err := GetLatestPlayerMapRecord(database, initial.SteamID, initial.KSFMapID)
	if err != nil {
		t.Fatalf("GetLatestPlayerMapRecord returned error: %v", err)
	}
	if latest == nil {
		t.Fatal("GetLatestPlayerMapRecord returned nil, want record")
	}
	if latest.SurfTimeMS != 104500 {
		t.Fatalf("latest SurfTimeMS = %d, want %d", latest.SurfTimeMS, 104500)
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM player_map_records`).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("record count = %d, want %d", count, 2)
	}
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	for _, path := range []string{
		migrationPath(t, "../../migrations/001_create_maps.sql"),
		migrationPath(t, "../../migrations/002_create_player_map_records.sql"),
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

func migrationPath(t *testing.T, relative string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), relative))
}
