package jobs

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/models"
)

func TestEnqueueMapRecordRefreshQueuesAvailableMap(t *testing.T) {
	database := openJobTestDatabase(t)
	seedMapRefreshJobData(t, database)

	now := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	result, err := enqueueMapRecordRefresh(database, 10, func() time.Time { return now })
	if err != nil {
		t.Fatalf("enqueueMapRecordRefresh returned error: %v", err)
	}
	if !result.Enqueued {
		t.Fatal("result.Enqueued = false, want true")
	}
	if result.Status != MapRecordRefreshStatusQueued {
		t.Fatalf("status = %q, want queued", result.Status)
	}
}

func TestEnqueueMapRecordRefreshReturnsExistingQueuedJob(t *testing.T) {
	database := openJobTestDatabase(t)
	seedMapRefreshJobData(t, database)

	now := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	if _, err := enqueueMapRecordRefresh(database, 10, func() time.Time { return now }); err != nil {
		t.Fatalf("initial enqueue failed: %v", err)
	}

	result, err := enqueueMapRecordRefresh(database, 10, func() time.Time { return now.Add(time.Minute) })
	if err != nil {
		t.Fatalf("duplicate enqueue returned error: %v", err)
	}
	if result.Enqueued {
		t.Fatal("duplicate result.Enqueued = true, want false")
	}
	if result.Status != MapRecordRefreshStatusQueued {
		t.Fatalf("status = %q, want queued", result.Status)
	}
}

func TestProcessNextMapRecordRefreshRefreshesAllStoredPlayersForMap(t *testing.T) {
	database := openJobTestDatabase(t)
	seedMapRefreshJobData(t, database)

	now := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	if _, err := enqueueMapRecordRefresh(database, 10, func() time.Time { return now }); err != nil {
		t.Fatalf("enqueueMapRecordRefresh returned error: %v", err)
	}

	scraper := mapRefreshStubScraper{
		recordsBySteamID: map[string]*models.PlayerMapRecord{
			"STEAM_0:1:1": {SteamID: "STEAM_0:1:1", PlayerID: 1, PlayerName: "Alpha", KSFMapID: 10, SurfTimeMS: 50000, Rank: 10, TotalRanks: 100},
			"STEAM_0:1:2": {SteamID: "STEAM_0:1:2", PlayerID: 2, PlayerName: "Bravo", KSFMapID: 10, SurfTimeMS: 60000, Rank: 20, TotalRanks: 100},
		},
	}

	processed, err := ProcessNextMapRecordRefresh(context.Background(), database, scraper, func() time.Time { return now.Add(time.Minute) })
	if err != nil {
		t.Fatalf("ProcessNextMapRecordRefresh returned error: %v", err)
	}
	if !processed {
		t.Fatal("processed = false, want true")
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM player_map_records WHERE ksf_map_id = 10`).Scan(&count); err != nil {
		t.Fatalf("count records failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("record count = %d, want 2", count)
	}

	state, err := db.GetMapRecordRefreshState(database, 10)
	if err != nil {
		t.Fatalf("GetMapRecordRefreshState returned error: %v", err)
	}
	if state == nil || state.Status != MapRecordRefreshStatusCompleted {
		t.Fatalf("refresh status = %+v, want completed", state)
	}
}

func TestProcessNextMapRecordRefreshRecordsPerPlayerFailures(t *testing.T) {
	database := openJobTestDatabase(t)
	seedMapRefreshJobData(t, database)

	if _, err := enqueueMapRecordRefresh(database, 10, time.Now); err != nil {
		t.Fatalf("enqueueMapRecordRefresh returned error: %v", err)
	}

	scraper := mapRefreshStubScraper{
		recordsBySteamID: map[string]*models.PlayerMapRecord{
			"STEAM_0:1:1": {SteamID: "STEAM_0:1:1", PlayerID: 1, PlayerName: "Alpha", KSFMapID: 10, SurfTimeMS: 50000, Rank: 10, TotalRanks: 100},
		},
		errorsBySteamID: map[string]error{
			"STEAM_0:1:2": errors.New("ksf unavailable"),
		},
	}

	processed, err := ProcessNextMapRecordRefresh(context.Background(), database, scraper, time.Now)
	if err != nil {
		t.Fatalf("ProcessNextMapRecordRefresh returned error: %v", err)
	}
	if !processed {
		t.Fatal("processed = false, want true")
	}

	state, err := db.GetMapRecordRefreshState(database, 10)
	if err != nil {
		t.Fatalf("GetMapRecordRefreshState returned error: %v", err)
	}
	if state == nil || state.Status != MapRecordRefreshStatusCompletedWithErrors {
		t.Fatalf("refresh status = %+v, want completed_with_errors", state)
	}
	if state.FailedPlayers != 1 {
		t.Fatalf("failed players = %d, want 1", state.FailedPlayers)
	}
}

func seedMapRefreshJobData(t *testing.T, database interface {
	Exec(query string, args ...any) (sql.Result, error)
}) {
	t.Helper()
	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES (10, 'surf_alpha', 1), (20, 'surf_beta', 2);
		INSERT INTO players (steam_id, player_id, name)
		VALUES ('STEAM_0:1:1', 1, 'Alpha'), ('STEAM_0:1:2', 2, 'Bravo');
	`); err != nil {
		t.Fatalf("seed data failed: %v", err)
	}
}

type mapRefreshStubScraper struct {
	recordsBySteamID map[string]*models.PlayerMapRecord
	errorsBySteamID  map[string]error
}

func (s mapRefreshStubScraper) FetchMainMapRecord(steamID string, _ string, _ int) (*models.PlayerMapRecord, error) {
	if err, ok := s.errorsBySteamID[steamID]; ok {
		return nil, err
	}
	record, ok := s.recordsBySteamID[steamID]
	if !ok {
		return nil, nil
	}
	return record, nil
}
