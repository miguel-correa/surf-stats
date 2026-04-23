package db

import (
	"os"
	"testing"

	"surfstats/internal/models"
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

	if _, err := database.Exec(`
		INSERT INTO player_map_records (steam_id, player_id, ksf_map_id, surf_time_ms, rank, total_ranks)
		VALUES
			('STEAM_0:1:100', 111, 1, 40000, 10, 100),
			('STEAM_0:1:100', 222, 2, 39000, 8, 100),
			('STEAM_0:1:200', 333, 3, 50000, 20, 100)
	`); err != nil {
		t.Fatalf("seed player_map_records failed: %v", err)
	}

	sqlBytes, err := os.ReadFile(migrationPath(t, "../../migrations/003_create_players.sql"))
	if err != nil {
		t.Fatalf("ReadFile(003_create_players.sql) failed: %v", err)
	}
	if _, err := database.Exec(string(sqlBytes)); err != nil {
		t.Fatalf("Exec migration 003_create_players.sql failed: %v", err)
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
