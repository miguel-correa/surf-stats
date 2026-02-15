CREATE TABLE
    IF NOT EXISTS maps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ksf_map_id INTEGER UNIQUE NOT NULL,
        name TEXT UNIQUE NOT NULL,
        tier INTEGER NOT NULL,
        added INTEGER,
        completions INTEGER DEFAULT 0,
        playtime_seconds INTEGER,
        comp_per_hour REAL DEFAULT 0,
        notes TEXT,
        bonus INTEGER DEFAULT 0,
        linear INTEGER DEFAULT 0,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );