package db

import (
	"testing"
	"testing/fstest"

	"surfstats/migrations"
)

func TestMigrateAppliesAllMigrations(t *testing.T) {
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count schema_migrations failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("schema_migrations count = %d, want 5", count)
	}

	for _, table := range []string{"maps", "player_map_records", "players", "map_record_refreshes"} {
		var name string
		err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s not found: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("first Migrate returned error: %v", err)
	}
	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count schema_migrations failed: %v", err)
	}
	if count != 5 {
		t.Fatalf("schema_migrations count = %d, want 5", count)
	}
}

func TestMigrateSkipsAppliedMigrations(t *testing.T) {
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate returned error: %v", err)
	}

	// Insert a row into maps to verify re-running doesn't drop/recreate the table
	if _, err := database.Exec("INSERT INTO maps (ksf_map_id, name, tier) VALUES (1, 'surf_test', 1)"); err != nil {
		t.Fatalf("insert test map failed: %v", err)
	}

	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("second Migrate returned error: %v", err)
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM maps").Scan(&count); err != nil {
		t.Fatalf("count maps failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("maps count = %d, want 1 (data should survive re-run)", count)
	}
}

func TestMigrateBadMigrationRollsBack(t *testing.T) {
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	fs := fstest.MapFS{
		"001_good.sql": &fstest.MapFile{Data: []byte("CREATE TABLE test_good (id INTEGER PRIMARY KEY);")},
		"002_bad.sql":  &fstest.MapFile{Data: []byte("INVALID SQL STATEMENT;")},
	}

	err := Migrate(database, fs)
	if err == nil {
		t.Fatal("Migrate should have returned error for bad SQL")
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = '001_good.sql'").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations failed: %v", err)
	}
	if count != 1 {
		t.Fatal("001_good.sql should be recorded in schema_migrations")
	}

	if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = '002_bad.sql'").Scan(&count); err != nil {
		t.Fatalf("query schema_migrations failed: %v", err)
	}
	if count != 0 {
		t.Fatal("002_bad.sql should NOT be recorded in schema_migrations")
	}
}
