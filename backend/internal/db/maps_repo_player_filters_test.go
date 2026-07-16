package db

import (
	"database/sql"
	"testing"
	"time"

	"surfstats/internal/models"
	"surfstats/migrations"
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
		SteamIDs:       []string{"STEAM_0:1:75949009"},
		PrimarySteamID: "STEAM_0:1:75949009",
		SortCol:        "id",
		Order:          "asc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionAll,
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
	if len(second.PlayerRecords) != 2 {
		t.Fatalf("len(second.PlayerRecords) = %d, want 2", len(second.PlayerRecords))
	}
	if !second.PlayerRecords[0].Completed {
		t.Fatal("second.PlayerRecords[0].Completed = false, want true")
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
		SteamID:    "STEAM_0:1:00000001",
		PlayerID:   111111,
		KSFMapID:   1,
		SurfTimeMS: 41000,
		Rank:       5000,
		TotalRanks: 27893,
	})
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 205000,
		Rank:       100,
		TotalRanks: 5000,
	})

	completed, err := GetMaps(database, MapFilters{
		SteamIDs:       []string{"STEAM_0:1:75949009", "STEAM_0:1:00000001"},
		PrimarySteamID: "STEAM_0:1:75949009",
		SortCol:        "id",
		Order:          "asc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionCompleted,
	})
	if err != nil {
		t.Fatalf("GetMaps completed returned error: %v", err)
	}
	if completed.Total != 1 || len(completed.Maps) != 1 {
		t.Fatalf("completed maps = total %d len %d, want 1/1", completed.Total, len(completed.Maps))
	}
	if completed.Maps[0].KSFMapID != 1 {
		t.Fatalf("completed map KSFMapID = %d, want 1", completed.Maps[0].KSFMapID)
	}

	incomplete, err := GetMaps(database, MapFilters{
		SteamIDs:       []string{"STEAM_0:1:75949009", "STEAM_0:1:00000001"},
		PrimarySteamID: "STEAM_0:1:75949009",
		SortCol:        "id",
		Order:          "asc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionIncomplete,
	})
	if err != nil {
		t.Fatalf("GetMaps incomplete returned error: %v", err)
	}
	if incomplete.Total != 1 || len(incomplete.Maps) != 1 {
		t.Fatalf("incomplete maps = total %d len %d, want 1/1", incomplete.Total, len(incomplete.Maps))
	}
	if incomplete.Maps[0].KSFMapID != 3 {
		t.Fatalf("incomplete map KSFMapID = %d, want 3", incomplete.Maps[0].KSFMapID)
	}
}

func TestGetMapsReturnsAllSelectedPlayerSummariesAndPrimaryProjection(t *testing.T) {
	database := openPlayerMapsTestDatabase(t)

	seedMapsForPlayerFilterTests(t, database)
	groupTier1 := 1
	groupTier3 := 3
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 105186,
		Rank:       142,
		TotalRanks: 6614,
		GroupTier:  &groupTier1,
	})
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:00000001",
		PlayerID:   111111,
		KSFMapID:   2,
		SurfTimeMS: 110500,
		Rank:       320,
		TotalRanks: 6614,
		GroupTier:  &groupTier3,
	})

	result, err := GetMaps(database, MapFilters{
		SteamIDs:       []string{"STEAM_0:1:75949009", "STEAM_0:1:00000001"},
		PrimarySteamID: "STEAM_0:1:00000001",
		SortCol:        "id",
		Order:          "asc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps returned error: %v", err)
	}

	second := result.Maps[1]
	if second.PlayerBestTimeMS == nil || *second.PlayerBestTimeMS != 110500 {
		t.Fatalf("primary PlayerBestTimeMS = %v, want 110500", second.PlayerBestTimeMS)
	}
	if len(second.PlayerRecords) != 2 {
		t.Fatalf("len(second.PlayerRecords) = %d, want 2", len(second.PlayerRecords))
	}
	if second.PlayerRecords[0].SteamID != "STEAM_0:1:75949009" {
		t.Fatalf("first player record steam id = %q, want STEAM_0:1:75949009", second.PlayerRecords[0].SteamID)
	}
	if second.PlayerRecords[1].SteamID != "STEAM_0:1:00000001" {
		t.Fatalf("second player record steam id = %q, want STEAM_0:1:00000001", second.PlayerRecords[1].SteamID)
	}
	if second.PlayerRecords[1].BestTimeMS == nil || *second.PlayerRecords[1].BestTimeMS != 110500 {
		t.Fatalf("second player record best time = %v, want 110500", second.PlayerRecords[1].BestTimeMS)
	}
	if second.PlayerRecords[0].GroupTier == nil || *second.PlayerRecords[0].GroupTier != 1 {
		t.Fatalf("first player record group tier = %v, want 1", second.PlayerRecords[0].GroupTier)
	}
	if second.PlayerRecords[1].GroupTier == nil || *second.PlayerRecords[1].GroupTier != 3 {
		t.Fatalf("second player record group tier = %v, want 3", second.PlayerRecords[1].GroupTier)
	}
	first := result.Maps[0]
	if len(first.PlayerRecords) != 2 {
		t.Fatalf("len(first.PlayerRecords) = %d, want 2", len(first.PlayerRecords))
	}
	if first.PlayerRecords[0].GroupTier != nil {
		t.Fatalf("incomplete player record group tier = %v, want nil", first.PlayerRecords[0].GroupTier)
	}
}

func TestGetMapsSortsByTier(t *testing.T) {
	database := openPlayerMapsTestDatabase(t)

	seedMapsForPlayerFilterTests(t, database)

	asc, err := GetMaps(database, MapFilters{
		SortCol:    "tier",
		Order:      "asc",
		Page:       1,
		PerPage:    10,
		Completion: CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps tier asc returned error: %v", err)
	}
	assertMapOrder(t, asc.Maps, []int{1, 2, 3})

	desc, err := GetMaps(database, MapFilters{
		SortCol:    "tier",
		Order:      "desc",
		Page:       1,
		PerPage:    10,
		Completion: CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps tier desc returned error: %v", err)
	}
	assertMapOrder(t, desc.Maps, []int{3, 2, 1})
}

func TestGetMapsSortsByPrimaryGroup(t *testing.T) {
	database := openPlayerMapsTestDatabase(t)

	seedMapsForPlayerFilterTests(t, database)
	group0 := 0
	group2 := 2
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   1,
		SurfTimeMS: 38565,
		Rank:       2306,
		TotalRanks: 27893,
		GroupTier:  &group2,
	})
	savePlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 205000,
		Rank:       100,
		TotalRanks: 5000,
		GroupTier:  &group0,
	})

	asc, err := GetMaps(database, MapFilters{
		PrimarySteamID: "STEAM_0:1:75949009",
		SortCol:        "primary_group",
		Order:          "asc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps primary_group asc returned error: %v", err)
	}
	assertMapOrder(t, asc.Maps, []int{2, 1, 3})

	desc, err := GetMaps(database, MapFilters{
		PrimarySteamID: "STEAM_0:1:75949009",
		SortCol:        "primary_group",
		Order:          "desc",
		Page:           1,
		PerPage:        10,
		Completion:     CompletionAll,
	})
	if err != nil {
		t.Fatalf("GetMaps primary_group desc returned error: %v", err)
	}
	assertMapOrder(t, desc.Maps, []int{3, 1, 2})
}

func openPlayerMapsTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })
	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
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

	_, err = database.Exec(`
		INSERT INTO players (steam_id, player_id, name)
		VALUES
			('STEAM_0:1:75949009', 949217, 'Player1'),
			('STEAM_0:1:00000001', 111111, 'Player2')
	`)
	if err != nil {
		t.Fatalf("seed players failed: %v", err)
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

func assertMapOrder(t *testing.T, maps []models.Map, want []int) {
	t.Helper()
	if len(maps) != len(want) {
		t.Fatalf("len(maps) = %d, want %d", len(maps), len(want))
	}
	for i, mapRef := range maps {
		if mapRef.KSFMapID != want[i] {
			t.Fatalf("maps[%d].KSFMapID = %d, want %d", i, mapRef.KSFMapID, want[i])
		}
	}
}
