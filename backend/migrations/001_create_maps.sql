CREATE TABLE IF NOT EXISTS MAPS (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    tier INTEGER NOT NULL,
    year INTEGER,
    completions INTEGER,
    hours_played INTEGER,
    comp_per_hour REAL,
    notes TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);