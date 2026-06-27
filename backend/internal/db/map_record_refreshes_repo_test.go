package db

import (
	"database/sql"
	"testing"
	"time"

	"surfstats/migrations"
)

func TestEnqueueMapRecordRefreshHonorsCooldown(t *testing.T) {
	database := openRefreshRepoTestDatabase(t)
	seedRefreshRepoMaps(t, database)

	start := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	first, err := EnqueueMapRecordRefresh(database, 1, start)
	if err != nil {
		t.Fatalf("first EnqueueMapRecordRefresh returned error: %v", err)
	}
	if !first.Enqueued {
		t.Fatal("first enqueue enqueued = false, want true")
	}

	if err := FinishMapRecordRefresh(database, 1, "completed", 0, 0, "", start.Add(time.Minute)); err != nil {
		t.Fatalf("FinishMapRecordRefresh returned error: %v", err)
	}

	second, err := EnqueueMapRecordRefresh(database, 1, start.Add(30*time.Second))
	if err != nil {
		t.Fatalf("second EnqueueMapRecordRefresh returned error: %v", err)
	}
	if !second.Cooldown {
		t.Fatal("second enqueue cooldown = false, want true")
	}
	if !second.AvailableAt.Equal(start.Add(time.Minute)) {
		t.Fatalf("second available at = %s, want %s", second.AvailableAt, start.Add(time.Minute))
	}

	third, err := EnqueueMapRecordRefresh(database, 1, start.Add(61*time.Second))
	if err != nil {
		t.Fatalf("third EnqueueMapRecordRefresh returned error: %v", err)
	}
	if !third.Enqueued {
		t.Fatal("third enqueue enqueued = false, want true")
	}
}

func TestEnqueueMapRecordRefreshReturnsExistingQueuedOrRunningJob(t *testing.T) {
	database := openRefreshRepoTestDatabase(t)
	seedRefreshRepoMaps(t, database)

	start := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	if _, err := EnqueueMapRecordRefresh(database, 1, start); err != nil {
		t.Fatalf("initial enqueue failed: %v", err)
	}

	queued, err := EnqueueMapRecordRefresh(database, 1, start.Add(time.Minute))
	if err != nil {
		t.Fatalf("queued duplicate returned error: %v", err)
	}
	if !queued.Existing || queued.State == nil || queued.State.Status != "queued" {
		t.Fatalf("queued duplicate = %+v, want existing queued", queued)
	}

	if _, err := ClaimNextQueuedMapRecordRefresh(database, start.Add(2*time.Minute)); err != nil {
		t.Fatalf("ClaimNextQueuedMapRecordRefresh returned error: %v", err)
	}

	running, err := EnqueueMapRecordRefresh(database, 1, start.Add(3*time.Minute))
	if err != nil {
		t.Fatalf("running duplicate returned error: %v", err)
	}
	if !running.Existing || running.State == nil || running.State.Status != "running" {
		t.Fatalf("running duplicate = %+v, want existing running", running)
	}
}

func TestClaimNextQueuedMapRecordRefreshClaimsOldestQueuedMap(t *testing.T) {
	database := openRefreshRepoTestDatabase(t)
	seedRefreshRepoMaps(t, database)

	start := time.Date(2026, time.May, 20, 12, 0, 0, 0, time.UTC)
	if _, err := EnqueueMapRecordRefresh(database, 2, start.Add(time.Minute)); err != nil {
		t.Fatalf("enqueue second map failed: %v", err)
	}
	if _, err := EnqueueMapRecordRefresh(database, 1, start); err != nil {
		t.Fatalf("enqueue first map failed: %v", err)
	}

	claimed, err := ClaimNextQueuedMapRecordRefresh(database, start.Add(2*time.Minute))
	if err != nil {
		t.Fatalf("ClaimNextQueuedMapRecordRefresh returned error: %v", err)
	}
	if claimed == nil || claimed.KSFMapID != 1 {
		t.Fatalf("claimed = %+v, want map 1", claimed)
	}

	state, err := GetMapRecordRefreshState(database, 1)
	if err != nil {
		t.Fatalf("GetMapRecordRefreshState returned error: %v", err)
	}
	if state == nil || state.Status != "running" {
		t.Fatalf("state = %+v, want running", state)
	}
}

func openRefreshRepoTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })
	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	return database
}

func seedRefreshRepoMaps(t *testing.T, database *sql.DB) {
	t.Helper()
	if _, err := database.Exec(`
		INSERT INTO maps (ksf_map_id, name, tier)
		VALUES (1, 'surf_alpha', 1), (2, 'surf_beta', 2)
	`); err != nil {
		t.Fatalf("seed maps failed: %v", err)
	}
}
