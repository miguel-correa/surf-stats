package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"surfstats/internal/models"
)

func GetMaps(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		allowedSortCols := map[string]string{
			"difficulty":  "comp_per_hour",
			"completions": "completions",
		}

		q := r.URL.Query()

		args := []any{}
		query := (`
			SELECT id, name, tier, year, completions, hours_played, comp_per_hour, notes
			FROM maps
			WHERE 1=1
		`)

		if tierStr := q.Get("tier"); tierStr != "" {
			tier, err := strconv.Atoi(tierStr)
			if err == nil && tier >= 1 && tier <= 8 {
				query += " AND tier = ?"
				args = append(args, tier)
			}
		}

		if search := q.Get("search"); search != "" {
			query += " AND name LIKE ?"
			args = append(args, "%"+search+"%")
		}

		sortKey := q.Get("sort")
		sortCol, ok := allowedSortCols[sortKey]
		if !ok {
			sortCol = allowedSortCols["difficulty"]
		}

		order := q.Get("order")
		if order != "asc" && order != "desc" {
			order = "desc"
		}

		query += " ORDER BY " + sortCol + " " + order

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		maps := []models.Map{}
		for rows.Next() {
			var m models.Map
			if err := rows.Scan(
				&m.ID,
				&m.Name,
				&m.Tier,
				&m.Year,
				&m.Completions,
				&m.HoursPlayed,
				&m.CompPerHour,
				&m.Notes,
			); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			maps = append(maps, m)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maps)
	}
}
