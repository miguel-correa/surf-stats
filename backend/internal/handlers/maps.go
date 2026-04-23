package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
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
		steamIDs := dedupeStrings(q["steam_id"])
		primarySteamID := q.Get("primary_steam_id")
		if primarySteamID == "" && len(steamIDs) > 0 {
			primarySteamID = steamIDs[0]
		}

		completion := db.CompletionAll
		switch q.Get("completion_status") {
		case string(db.CompletionCompleted):
			completion = db.CompletionCompleted
		case string(db.CompletionIncomplete):
			completion = db.CompletionIncomplete
		}

		var linearFilter *int
		linear := q.Get("linear")
		if linear == "0" || linear == "1" {
			v, err := strconv.Atoi(linear)
			if err == nil {
				linearFilter = &v
			}
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

		page, _ := strconv.Atoi(q.Get("page"))
		if page < 1 {
			page = 1
		}

		perPage, _ := strconv.Atoi(q.Get("per_page"))
		if perPage < 1 || perPage > 100 {
			perPage = 10
		}

		filters := db.MapFilters{
			Tiers:          tiers,
			Search:         search,
			Linear:         linearFilter,
			SteamIDs:       steamIDs,
			PrimarySteamID: primarySteamID,
			Completion:     completion,
			SortCol:        sortCol,
			Order:          order,
			Page:           page,
			PerPage:        perPage,
		}

		maps, err := db.GetMaps(database, filters)
		if err != nil {
			log.Printf("GetMaps error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(maps)
	}
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
