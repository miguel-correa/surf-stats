CREATE TABLE
    IF NOT EXISTS player_map_records (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        steam_id TEXT NOT NULL,
        player_id INTEGER,
        ksf_map_id INTEGER NOT NULL,
        surf_time_ms INTEGER NOT NULL,
        rank INTEGER,
        total_ranks INTEGER,
        date_set TIMESTAMP,
        completions INTEGER,
        group_tier INTEGER,
        scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

CREATE INDEX IF NOT EXISTS idx_player_map_records_player_map_scraped
    ON player_map_records (steam_id, ksf_map_id, scraped_at DESC);

CREATE INDEX IF NOT EXISTS idx_player_map_records_player_map_time
    ON player_map_records (steam_id, ksf_map_id, surf_time_ms);

