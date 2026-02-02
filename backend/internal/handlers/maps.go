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

		var tiers []int
		tierStrs := q["tier"]
		for _, t := range tierStrs {
			tier, err := strconv.Atoi(t)
			if err == nil && tier >= 1 && tier <= 8 {
				tiers = append(tiers, tier)
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

		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}

		perPage, _ := strconv.Atoi(q.Get("per_page"))
		if perPage < 1 || perPage > 100 {
			perPage = 10
		}

		filters := db.MapFilters{
			Tiers:   tiers,
			Search:  search,
			SortCol: sortCol,
			Order:   order,
			Page:    page,
			PerPage: perPage,
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
