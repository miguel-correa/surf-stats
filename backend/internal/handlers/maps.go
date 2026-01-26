package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"surfstats/internal/models"
)

func GetMaps(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`
			SELECT id, name, tier, year, completions, hours_played, comp_per_hour, notes
			FROM maps
			ORDER BY comp_per_hour DESC
		`)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		defer rows.Close()

		var maps []models.Map
		for rows.Next() {
			var m models.Map
			rows.Scan(
				&m.ID,
				&m.Name,
				&m.Tier,
				&m.Year,
				&m.Completions,
				&m.HoursPlayed,
				&m.CompPerHour,
				&m.Notes,
			)
			maps = append(maps, m)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maps)
	}
}
