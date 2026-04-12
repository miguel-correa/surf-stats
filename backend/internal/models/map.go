package models

import "time"

type Map struct {
	ID               int       `json:"id"`
	KSFMapID         int       `json:"ksf_map_id"`
	Name             string    `json:"name"`
	Tier             int       `json:"tier"`
	Added            int       `json:"added"`
	Completions      int       `json:"completions"`
	PlaytimeSeconds  int       `json:"playtime_seconds"`
	CompPerHour      float64   `json:"comp_per_hour"`
	Notes            string    `json:"notes"`
	Bonus            int       `json:"bonus"`
	Linear           int       `json:"linear"`
	UpdatedAt        time.Time `json:"updated_at"`
	PlayerBestTimeMS *int      `json:"player_best_time_ms"`
	PlayerRank       *int      `json:"player_rank"`
	Completed        *bool     `json:"completed"`
}
