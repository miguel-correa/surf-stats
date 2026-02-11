package models

import "time"

type Map struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Tier        int     `json:"tier"`
	Year        int     `json:"year"`
	Completions int     `json:"completions"`
	HoursPlayed float64 `json:"hours_played"`
	CompPerHour float64 `json:"comp_per_hour"`
	Notes       string  `json:"notes"`
}

type MapV2 struct {
	ID              int       `json:"id"`
	KSFMapID        int       `json:"ksf_map_id"`
	Name            string    `json:"name"`
	Tier            int       `json:"tier"`
	Added           int       `json:"year"`
	Completions     int       `json:"completions"`
	PlaytimeSeconds int       `json:"playtime_seconds"`
	CompPerHour     float64   `json:"comp_per_hour"`
	Notes           string    `json:"notes"`
	Bonus           int       `json:"bonus"`
	Linear          int       `json:"linear"`
	UpdatedAt       time.Time `json:"updated_at"`
}
