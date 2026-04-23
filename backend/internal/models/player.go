package models

type Player struct {
	SteamID  string `json:"steam_id"`
	PlayerID *int   `json:"player_id"`
	Name     string `json:"name"`
}
