package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"surfstats/internal/db"
	"surfstats/migrations"
)

func TestGetPlayersReturnsPlayersJSON(t *testing.T) {
	database := db.Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })

	if err := db.Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
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

