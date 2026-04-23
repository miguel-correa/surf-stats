package db

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"surfstats/internal/models"
	"surfstats/migrations"
)

func TestListPlayersReturnsAlphabeticalPlayers(t *testing.T) {
	database := openTestDatabase(t)

	seedPlayer := func(steamID string, playerID int, name string) {
		t.Helper()

		playerIDCopy := playerID
		if err := UpsertPlayer(database, models.Player{
			SteamID:  steamID,
			PlayerID: &playerIDCopy,
			Name:     name,
		}); err != nil {
			t.Fatalf("UpsertPlayer(%q) failed: %v", steamID, err)
		}
	}

	seedPlayer("STEAM_0:1:00000002", 2, "Zulu")
	seedPlayer("STEAM_0:1:00000001", 1, "Alpha")

	players, err := ListPlayers(database)
	if err != nil {
		t.Fatalf("ListPlayers returned error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("len(players) = %d, want 2", len(players))
	}
	if players[0].Name != "Alpha" || players[1].Name != "Zulu" {
		t.Fatalf("player order = [%q, %q], want [Alpha, Zulu]", players[0].Name, players[1].Name)
	}
}

func TestPlayerMigrationBackfillsDistinctSteamIDs(t *testing.T) {
	database := Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	// Apply only migrations 001 and 002 using a partial FS
	partialFS := make(fstest.MapFS)
	for _, name := range []string{"001_create_maps.sql", "002_create_player_map_records.sql"} {
		data, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			t.Fatalf("read %s from embed: %v", name, err)
		}
		partialFS[name] = &fstest.MapFile{Data: data}
	}
	if err := Migrate(database, partialFS); err != nil {
		t.Fatalf("Migrate (partial) failed: %v", err)
	}

	if _, err := database.Exec(`
		INSERT INTO player_map_records (steam_id, player_id, ksf_map_id, surf_time_ms, rank, total_ranks)
		VALUES
			('STEAM_0:1:100', 111, 1, 40000, 10, 100),
			('STEAM_0:1:100', 222, 2, 39000, 8, 100),
			('STEAM_0:1:200', 333, 3, 50000, 20, 100)
	`); err != nil {
		t.Fatalf("seed player_map_records failed: %v", err)
	}

	// Now apply the full FS — only 003 should run, triggering the backfill
	if err := Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate (full) failed: %v", err)
	}

	players, err := ListPlayers(database)
	if err != nil {
		t.Fatalf("ListPlayers returned error: %v", err)
	}
	if len(players) != 2 {
		t.Fatalf("len(players) = %d, want 2", len(players))
	}
	if players[0].SteamID != "STEAM_0:1:100" {
		t.Fatalf("players[0].SteamID = %q, want STEAM_0:1:100", players[0].SteamID)
	}
	if players[0].PlayerID == nil || *players[0].PlayerID != 222 {
		t.Fatalf("players[0].PlayerID = %v, want 222", players[0].PlayerID)
	}
	if players[0].Name != "STEAM_0:1:100" {
		t.Fatalf("players[0].Name = %q, want steam_id fallback", players[0].Name)
	}
}
