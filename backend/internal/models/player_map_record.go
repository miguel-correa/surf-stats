package models

import "time"

type PlayerMapRecord struct {
	ID          int        `json:"id"`
	SteamID     string     `json:"steam_id"`
	PlayerID    int        `json:"player_id"`
	KSFMapID    int        `json:"ksf_map_id"`
	SurfTimeMS  int        `json:"surf_time_ms"`
	Rank        int        `json:"rank"`
	TotalRanks  int        `json:"total_ranks"`
	DateSet     *time.Time `json:"date_set"`
	Completions *int       `json:"completions"`
	GroupTier   *int       `json:"group_tier"`
	ScrapedAt   time.Time  `json:"scraped_at"`
}
