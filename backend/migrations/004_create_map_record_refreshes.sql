CREATE TABLE
    IF NOT EXISTS map_record_refreshes (
        ksf_map_id INTEGER PRIMARY KEY,
        last_started_at TIMESTAMP NOT NULL,
        last_finished_at TIMESTAMP,
        status TEXT NOT NULL,
        refreshed_players INTEGER DEFAULT 0,
        failed_players INTEGER DEFAULT 0,
        last_error TEXT,
        FOREIGN KEY (ksf_map_id) REFERENCES maps (ksf_map_id)
    );
