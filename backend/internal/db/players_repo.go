package db

import (
	"database/sql"
	"fmt"
	"strings"

	"surfstats/internal/models"
)

func ListPlayers(database *sql.DB) ([]models.Player, error) {
	rows, err := database.Query(`
		SELECT steam_id, player_id, name
		FROM players
		ORDER BY name COLLATE NOCASE ASC, steam_id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var player models.Player
		var playerID sql.NullInt64
		if err := rows.Scan(&player.SteamID, &playerID, &player.Name); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		if playerID.Valid {
			value := int(playerID.Int64)
			player.PlayerID = &value
		}
		players = append(players, player)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate players: %w", err)
	}

	return players, nil
}

func UpsertPlayer(database *sql.DB, player models.Player) error {
	name := strings.TrimSpace(player.Name)
	if name == "" {
		name = player.SteamID
	}

	_, err := database.Exec(`
		INSERT INTO players (steam_id, player_id, name)
		VALUES (?, ?, ?)
		ON CONFLICT(steam_id) DO UPDATE SET
			player_id = excluded.player_id,
			name = excluded.name,
			updated_at = CURRENT_TIMESTAMP
	`, player.SteamID, player.PlayerID, name)
	if err != nil {
		return fmt.Errorf("upsert player steam_id=%s: %w", player.SteamID, err)
	}
	return nil
}
