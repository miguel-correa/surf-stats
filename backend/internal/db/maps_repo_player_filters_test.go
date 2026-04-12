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

func TestGetMapsIncludesPlayerRecordFields(t *testing.T) {
	database := openPlayerMapsTestDatabase(t)

	seedMapsForPlayerFilterTests(t, database)
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 105186,
		Rank:       142,
		TotalRanks: 6614,
	})

	result, err := GetMaps(database, MapFilters{
		SteamID:    "STEAM_0:1:75949009",
		SortCol:    "id",
		Order:      "asc",
		Page:       1,
		PerPage:    10,
		Completion: CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps returned error: %v", err)
	}

	if len(result.Maps) != 3 {
		t.Fatalf("len(result.Maps) = %d, want 3", len(result.Maps))
	}

	first := result.Maps[0]
	if first.Completed == nil || *first.Completed {
		t.Fatalf("first.Completed = %v, want false", first.Completed)
	}
	if first.PlayerBestTimeMS != nil {
		t.Fatalf("first.PlayerBestTimeMS = %v, want nil", first.PlayerBestTimeMS)
	}

	second := result.Maps[1]
	if second.Completed == nil || !*second.Completed {
		t.Fatalf("second.Completed = %v, want true", second.Completed)
	}
	if second.PlayerBestTimeMS == nil || *second.PlayerBestTimeMS != 105186 {
		t.Fatalf("second.PlayerBestTimeMS = %v, want 105186", second.PlayerBestTimeMS)
	}
	if second.PlayerRank == nil || *second.PlayerRank != 142 {
		t.Fatalf("second.PlayerRank = %v, want 142", second.PlayerRank)
	}
}

func TestGetMapsFiltersByCompletionStatus(t *testing.T) {
	database := openPlayerMapsTestDatabase(t)

	seedMapsForPlayerFilterTests(t, database)
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   1,
		SurfTimeMS: 38565,
		Rank:       2306,
		TotalRanks: 27893,
	})
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   3,
		SurfTimeMS: 205000,
		Rank:       100,
		TotalRanks: 5000,
	})

	completed, err := GetMaps(database, MapFilters{
		SteamID:    "STEAM_0:1:75949009",
		SortCol:    "id",
		Order:      "asc",
		Page:       1,
		PerPage:    10,
		Completion: CompletionCompleted,
	})
	if err != nil {
		t.Fatalf("GetMaps completed returned error: %v", err)
	}
	if completed.Total != 2 || len(completed.Maps) != 2 {
		t.Fatalf("completed maps = total %d len %d, want 2/2", completed.Total, len(completed.Maps))
	}

	incomplete, err := GetMaps(database, MapFilters{
		SteamID:    "STEAM_0:1:75949009",
		SortCol:    "id",
		Order:      "asc",
		Page:       1,
		PerPage:    10,
		Completion: CompletionIncomplete,
	})
	if err != nil {
		t.Fatalf("GetMaps incomplete returned error: %v", err)
	}
	if incomplete.Total != 1 || len(incomplete.Maps) != 1 {
		t.Fatalf("incomplete maps = total %d len %d, want 1/1", incomplete.Total, len(incomplete.Maps))
	}
	if incomplete.Maps[0].KSFMapID != 2 {
		t.Fatalf("incomplete map KSFMapID = %d, want 2", incomplete.Maps[0].KSFMapID)
	}
}

func openPlayerMapsTestDatabase(t *testing.T) *sql.DB {
	t.Helper()

	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	for _, path := range []string{
		playerMapsMigrationPath(t, "../../migrations/001_create_maps.sql"),
		playerMapsMigrationPath(t, "../../migrations/002_create_player_map_records.sql"),
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

func seedMapsForPlayerFilterTests(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier, added, completions, playtime_seconds, comp_per_hour, bonus, linear)
		VALUES
			(1, 'surf_alpha', 1, 1710000000, 100, 3600, 100.0, 0, 1),
			(2, 'surf_beta', 2, 1710000100, 200, 7200, 100.0, 0, 0),
			(3, 'surf_gamma', 3, 1710000200, 300, 10800, 100.0, 0, 1)
	`)
	if err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}
}

func savePlayerPB(t *testing.T, database *sql.DB, record models.PlayerMapRecord) {
	t.Helper()

	dateSet := time.Date(2026, time.April, 10, 19, 13, 42, 0, time.UTC)
	record.DateSet = &dateSet

	inserted, err := SavePlayerMapRecordIfImproved(database, record)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved failed: %v", err)
	}
	if !inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = false, want true")
	}
}

func playerMapsMigrationPath(t *testing.T, relative string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), relative))
}
