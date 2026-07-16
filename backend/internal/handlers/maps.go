package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
	"time"
)

func GetMaps(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedSortCols := map[string]string{
			"difficulty":    "comp_per_hour",
			"completions":   "completions",
			"tier":          "tier",
			"primary_group": "primary_group",
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

type mapRecordRefreshRunner func(*sql.DB, int) (jobs.MapRecordRefreshResult, error)

type mapRecordRefreshResponse struct {
	Status           string `json:"status,omitempty"`
	RefreshedPlayers int    `json:"refreshed_players,omitempty"`
	FailedPlayers    int    `json:"failed_players,omitempty"`
	AvailableAt      string `json:"available_at,omitempty"`
	Error            string `json:"error,omitempty"`
}

func RefreshMapRecords(database *sql.DB) http.HandlerFunc {
	return refreshMapRecords(database, jobs.EnqueueMapRecordRefresh)
}

func refreshMapRecords(database *sql.DB, runner mapRecordRefreshRunner) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ksfMapID, ok := parseRefreshMapID(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		result, err := runner(database, ksfMapID)
		if err != nil {
			if cooldownErr, ok := jobs.IsMapRecordRefreshCooldown(err); ok {
				availableAt := cooldownErr.AvailableAt.UTC()
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(int(timeUntilSeconds(availableAt))))
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(mapRecordRefreshResponse{
					AvailableAt: availableAt.Format(time.RFC3339),
					Error:       "refresh cooldown active",
				})
				return
			}
			if jobs.IsMapRecordRefreshNotFound(err) {
				http.NotFound(w, r)
				return
			}

			log.Printf("RefreshMapRecords error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		statusCode := http.StatusOK
		if result.Enqueued {
			statusCode = http.StatusAccepted
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(mapRecordRefreshResponse{
			Status:           result.Status,
			RefreshedPlayers: result.RefreshedPlayers,
			FailedPlayers:    result.FailedPlayers,
			AvailableAt:      formatOptionalTime(result.AvailableAt),
		})
	}
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func parseRefreshMapID(path string) (int, bool) {
	const prefix = "/api/maps/"
	const suffix = "/refresh-records"
	if !strings.HasPrefix(path, prefix) || !strings.HasSuffix(path, suffix) {
		return 0, false
	}

	rawID := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	rawID = strings.Trim(rawID, "/")
	if rawID == "" || strings.Contains(rawID, "/") {
		return 0, false
	}

	ksfMapID, err := strconv.Atoi(rawID)
	if err != nil || ksfMapID < 1 {
		return 0, false
	}
	return ksfMapID, true
}

func timeUntilSeconds(t time.Time) int64 {
	seconds := time.Until(t).Seconds()
	if seconds < 0 {
		return 0
	}
	return int64(seconds)
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
