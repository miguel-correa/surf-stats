package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"surfstats/internal/db"
)

func GetMaps(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		allowedSortCols := map[string]string{
			"difficulty":  "comp_per_hour",
			"completions": "completions",
		}

		q := r.URL.Query()

		var tierPtr *int
		if tierStr := q.Get("tier"); tierStr != "" {
			tier, err := strconv.Atoi(tierStr)
			if err == nil && tier >= 1 && tier <= 8 {
				tierPtr = &tier
			}
		}

		search := q.Get("search")

		sortKey := q.Get("sort")
		sortCol, ok := allowedSortCols[sortKey]
		if !ok {
			sortCol = allowedSortCols["difficulty"]
		}

		order := q.Get("order")
		if order != "asc" && order != "desc" {
			order = "desc"
		}

		filters := db.MapFilters{
			Tier:    tierPtr,
			Search:  search,
			SortCol: sortCol,
			Order:   order,
		}

		maps, err := db.GetMaps(database, filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maps)
	}
}
