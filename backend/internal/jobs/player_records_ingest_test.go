package jobs

import (
	"database/sql"
	"sync"
	"testing"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/players"
	"surfstats/migrations"
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
				PlayerName: "Billy Rubina",
				KSFMapID:   1,
				SurfTimeMS: 45000,
				Rank:       50,
				TotalRanks: 500,
			},
			"surf_beta": {
				SteamID:    "STEAM_0:1:75949009",
				PlayerID:   949217,
				PlayerName: "Billy Rubina",
				KSFMapID:   2,
				SurfTimeMS: 59000,
				Rank:       90,
				TotalRanks: 1000,
			},
			"surf_gamma": nil,
		},
	}
	installPlayerUpsertAudit(t, database)

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

	players, err := db.ListPlayers(database)
	if err != nil {
		t.Fatalf("ListPlayers returned error: %v", err)
	}
	if len(players) != 1 {
		t.Fatalf("len(players) = %d, want 1", len(players))
	}
	if players[0].Name != "Billy Rubina" {
		t.Fatalf("players[0].Name = %q, want Billy Rubina", players[0].Name)
	}
	upsertCount := playerUpsertAuditCount(t, database)
	if upsertCount != 1 {
		t.Fatalf("player upsert count = %d, want 1", upsertCount)
	}
}

func TestRunPlayerRecordsIngestionFetchesMapsConcurrently(t *testing.T) {
	database := openJobTestDatabase(t)

	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES
			(1, 'surf_alpha', 1),
			(2, 'surf_beta', 2),
			(3, 'surf_gamma', 3),
			(4, 'surf_delta', 4)
	`); err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}

	scraper := &blockingMainRecordScraper{
		recordsByMapName: map[string]*models.PlayerMapRecord{
			"surf_alpha": {SteamID: "STEAM_0:1:75949009", PlayerID: 1, PlayerName: "Billy Rubina", KSFMapID: 1, SurfTimeMS: 1000, Rank: 1, TotalRanks: 10},
			"surf_beta":  {SteamID: "STEAM_0:1:75949009", PlayerID: 1, PlayerName: "Billy Rubina", KSFMapID: 2, SurfTimeMS: 1000, Rank: 1, TotalRanks: 10},
			"surf_gamma": {SteamID: "STEAM_0:1:75949009", PlayerID: 1, PlayerName: "Billy Rubina", KSFMapID: 3, SurfTimeMS: 1000, Rank: 1, TotalRanks: 10},
			"surf_delta": {SteamID: "STEAM_0:1:75949009", PlayerID: 1, PlayerName: "Billy Rubina", KSFMapID: 4, SurfTimeMS: 1000, Rank: 1, TotalRanks: 10},
		},
		releaseCh: make(chan struct{}),
	}

	done := make(chan error, 1)
	go func() {
		done <- runPlayerRecordsIngestionWithWorkers(database, "STEAM_0:1:75949009", scraper, 4)
	}()

	deadline := time.After(2 * time.Second)
	for {
		scraper.mu.Lock()
		maxConcurrent := scraper.maxConcurrent
		scraper.mu.Unlock()
		if maxConcurrent >= 2 {
			break
		}

		select {
		case err := <-done:
			t.Fatalf("runPlayerRecordsIngestionWithWorkers finished too early: %v", err)
		case <-deadline:
			t.Fatalf("max concurrency never exceeded 1; got %d", maxConcurrent)
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}

	close(scraper.releaseCh)

	if err := <-done; err != nil {
		t.Fatalf("runPlayerRecordsIngestionWithWorkers returned error: %v", err)
	}
}

func TestRunPlayerRecordsIngestionSkipsPerMapFailures(t *testing.T) {
	database := openJobTestDatabase(t)

	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES
			(1, 'surf_alpha', 1),
			(2, 'surf_beta', 2)
	`); err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}

	scraper := stubMainRecordScraper{
		recordsByMapName: map[string]*models.PlayerMapRecord{
			"surf_alpha": {
				SteamID:    "STEAM_0:1:75949009",
				PlayerID:   949217,
				PlayerName: "Billy Rubina",
				KSFMapID:   1,
				SurfTimeMS: 45000,
				Rank:       50,
				TotalRanks: 500,
			},
		},
		errorsByMapName: map[string]error{
			"surf_beta": players.HTTPStatusError{StatusCode: 500},
		},
	}

	if err := runPlayerRecordsIngestion(database, "STEAM_0:1:75949009", scraper); err != nil {
		t.Fatalf("runPlayerRecordsIngestion returned error: %v", err)
	}

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM player_map_records`).Scan(&count); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("record count = %d, want %d", count, 1)
	}
}

func TestRunPlayerRecordsIngestionRetriesTransientStatusErrors(t *testing.T) {
	database := openJobTestDatabase(t)

	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES (1, 'surf_alpha', 1)
	`); err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}

	scraper := &flakyMainRecordScraper{
		failuresBeforeSuccess: map[string]int{
			"surf_alpha": 2,
		},
		record: &models.PlayerMapRecord{
			SteamID:    "STEAM_0:1:75949009",
			PlayerID:   949217,
			PlayerName: "Billy Rubina",
			KSFMapID:   1,
			SurfTimeMS: 45000,
			Rank:       50,
			TotalRanks: 500,
		},
	}

	if err := runPlayerRecordsIngestionWithWorkers(database, "STEAM_0:1:75949009", scraper, 1); err != nil {
		t.Fatalf("runPlayerRecordsIngestionWithWorkers returned error: %v", err)
	}

	if scraper.calls != 3 {
		t.Fatalf("scraper calls = %d, want 3", scraper.calls)
	}
}

type stubMainRecordScraper struct {
	recordsByMapName map[string]*models.PlayerMapRecord
	errorsByMapName  map[string]error
}

func (s stubMainRecordScraper) FetchMainMapRecord(_ string, mapName string, _ int) (*models.PlayerMapRecord, error) {
	if err, ok := s.errorsByMapName[mapName]; ok {
		return nil, err
	}
	record, ok := s.recordsByMapName[mapName]
	if !ok {
		return nil, nil
	}
	return record, nil
}

type flakyMainRecordScraper struct {
	failuresBeforeSuccess map[string]int
	record                *models.PlayerMapRecord
	calls                 int
}

func (s *flakyMainRecordScraper) FetchMainMapRecord(_ string, mapName string, _ int) (*models.PlayerMapRecord, error) {
	s.calls++
	if s.failuresBeforeSuccess[mapName] > 0 {
		s.failuresBeforeSuccess[mapName]--
		return nil, players.HTTPStatusError{StatusCode: 500}
	}
	return s.record, nil
}

type blockingMainRecordScraper struct {
	recordsByMapName map[string]*models.PlayerMapRecord
	releaseCh        chan struct{}
	mu               sync.Mutex
	inFlight         int
	maxConcurrent    int
}

func (s *blockingMainRecordScraper) FetchMainMapRecord(_ string, mapName string, _ int) (*models.PlayerMapRecord, error) {
	s.mu.Lock()
	s.inFlight++
	if s.inFlight > s.maxConcurrent {
		s.maxConcurrent = s.inFlight
	}
	s.mu.Unlock()

	<-s.releaseCh

	s.mu.Lock()
	s.inFlight--
	s.mu.Unlock()

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
	if err := db.Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	return database
}

func installPlayerUpsertAudit(t *testing.T, database *sql.DB) {
	t.Helper()

	_, err := database.Exec(`
		CREATE TABLE player_upsert_audit (action TEXT NOT NULL);
		CREATE TRIGGER players_insert_audit
		AFTER INSERT ON players
		BEGIN
			INSERT INTO player_upsert_audit (action) VALUES ('insert');
		END;
		CREATE TRIGGER players_update_audit
		AFTER UPDATE ON players
		BEGIN
			INSERT INTO player_upsert_audit (action) VALUES ('update');
		END;
	`)
	if err != nil {
		t.Fatalf("install player upsert audit failed: %v", err)
	}
}

func playerUpsertAuditCount(t *testing.T, database *sql.DB) int {
	t.Helper()

	var count int
	if err := database.QueryRow(`SELECT COUNT(*) FROM player_upsert_audit`).Scan(&count); err != nil {
		t.Fatalf("player upsert audit count query failed: %v", err)
	}
	return count
}
