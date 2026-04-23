CREATE TABLE
    IF NOT EXISTS players (
        steam_id TEXT PRIMARY KEY,
        player_id INTEGER,
        name TEXT NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

INSERT INTO players (steam_id, player_id, name)
SELECT pr.steam_id, pr.player_id, pr.steam_id
FROM player_map_records pr
JOIN (
    SELECT steam_id, MAX(id) AS max_id
    FROM player_map_records
    GROUP BY steam_id
) latest ON latest.max_id = pr.id
ON CONFLICT(steam_id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_players_name
    ON players (name);
