package models

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
