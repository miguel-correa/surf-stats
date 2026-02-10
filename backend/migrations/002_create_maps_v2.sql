CREATE TABLE
    IF NOT EXISTS maps_v2 (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ksf_map_id INTEGER UNIQUE NOT NULL,
        name TEXT UNIQUE NOT NULL,
        tier INTEGER NOT NULL,
        year INTEGER,
        completions INTEGER, -- from totalRanks (main zone)
        playtime_seconds INTEGER, -- raw seconds from scrape
        comp_per_hour REAL,
        notes TEXT,
        bonus INTEGER DEFAULT 0, -- from b_count
        linear INTEGER DEFAULT 0, -- from isLinear (0/1)
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );