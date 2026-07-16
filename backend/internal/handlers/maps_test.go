package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"surfstats/internal/db"
	"surfstats/internal/models"
)

func TestGetMapsAcceptsTierSort(t *testing.T) {
	database := openHandlerTestDatabase(t)

	seedMapsForSortHandlerTests(t, database)

	req := httptest.NewRequest(http.MethodGet, "/api/maps?sort=tier&order=asc&per_page=10", nil)
	rec := httptest.NewRecorder()

	GetMaps(database).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	result := decodeMapsResponse(t, rec)
	assertHandlerMapOrder(t, result.Maps, []int{2, 1, 3})
}

func TestGetMapsAcceptsPrimaryGroupSort(t *testing.T) {
	database := openHandlerTestDatabase(t)

	seedMapsForSortHandlerTests(t, database)
	group0 := 0
	group2 := 2
	saveHandlerPlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   1,
		SurfTimeMS: 38565,
		Rank:       2306,
		TotalRanks: 27893,
		GroupTier:  &group2,
	})
	saveHandlerPlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 205000,
		Rank:       100,
		TotalRanks: 5000,
		GroupTier:  &group0,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/maps?primary_steam_id=STEAM_0:1:75949009&sort=primary_group&order=asc&per_page=10", nil)
	rec := httptest.NewRecorder()

	GetMaps(database).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	result := decodeMapsResponse(t, rec)
	assertHandlerMapOrder(t, result.Maps, []int{2, 1, 3})
}

func TestGetMapsAcceptsPrimaryGroupFilter(t *testing.T) {
	database := openHandlerTestDatabase(t)

	seedMapsForSortHandlerTests(t, database)
	group5 := 5
	group6 := 6
	saveHandlerPlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   1,
		SurfTimeMS: 38565,
		Rank:       2306,
		TotalRanks: 27893,
		GroupTier:  &group6,
	})
	saveHandlerPlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   2,
		SurfTimeMS: 205000,
		Rank:       100,
		TotalRanks: 5000,
		GroupTier:  &group5,
	})
	saveHandlerPlayerPB(t, database, models.PlayerMapRecord{
		SteamID:    "STEAM_0:1:75949009",
		PlayerID:   949217,
		KSFMapID:   3,
		SurfTimeMS: 155000,
		Rank:       300,
		TotalRanks: 5000,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/maps?primary_steam_id=STEAM_0:1:75949009&primary_group=6&primary_group=5&primary_group=bad&primary_group=7&sort=tier&order=asc&per_page=10", nil)
	rec := httptest.NewRecorder()

	GetMaps(database).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}

	result := decodeMapsResponse(t, rec)
	assertHandlerMapOrder(t, result.Maps, []int{2, 1})
}

func seedMapsForSortHandlerTests(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier, added, completions, playtime_seconds, comp_per_hour, bonus, linear)
		VALUES
			(1, 'surf_alpha', 2, 1710000000, 100, 1000, 3.0, 0, 1),
			(2, 'surf_beta', 1, 1710000100, 200, 1000, 2.0, 0, 1),
			(3, 'surf_gamma', 3, 1710000200, 300, 1000, 1.0, 0, 1)
	`)
	if err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}
}

func saveHandlerPlayerPB(t *testing.T, database *sql.DB, record models.PlayerMapRecord) {
	t.Helper()

	inserted, err := db.SavePlayerMapRecordIfImproved(database, record)
	if err != nil {
		t.Fatalf("SavePlayerMapRecordIfImproved failed: %v", err)
	}
	if !inserted {
		t.Fatal("SavePlayerMapRecordIfImproved inserted = false, want true")
	}
}

func decodeMapsResponse(t *testing.T, rec *httptest.ResponseRecorder) db.PaginatedMaps {
	t.Helper()

	var result db.PaginatedMaps
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	return result
}

func assertHandlerMapOrder(t *testing.T, maps []models.Map, want []int) {
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
