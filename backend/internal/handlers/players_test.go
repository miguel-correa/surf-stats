package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"surfstats/internal/db"
)

func TestGetPlayersReturnsPlayersJSON(t *testing.T) {
	database := db.Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	for _, relative := range []string{
		"../../migrations/002_create_player_map_records.sql",
		"../../migrations/003_create_players.sql",
	} {
		sqlBytes, err := os.ReadFile(playersMigrationPath(t, relative))
		if err != nil {
			t.Fatalf("ReadFile(%s) failed: %v", relative, err)
		}
		if _, err := database.Exec(string(sqlBytes)); err != nil {
			t.Fatalf("Exec migration %s failed: %v", relative, err)
		}
	}
	if _, err := database.Exec(`
		INSERT INTO players (steam_id, player_id, name)
		VALUES ('STEAM_0:1:1', 1, 'Alpha'), ('STEAM_0:1:2', 2, 'Bravo')
	`); err != nil {
		t.Fatalf("seed players failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/players", nil)
	rec := httptest.NewRecorder()

	GetPlayers(database).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	want := "[{\"steam_id\":\"STEAM_0:1:1\",\"player_id\":1,\"name\":\"Alpha\"},{\"steam_id\":\"STEAM_0:1:2\",\"player_id\":2,\"name\":\"Bravo\"}]\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %q, want %q", rec.Body.String(), want)
	}
}

func playersMigrationPath(t *testing.T, relative string) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), relative))
}
