package db

import (
	"database/sql"
	"fmt"
	"strings"

	"surfstats/internal/models"
	"surfstats/internal/scrapers/maps"
)

type MapFilters struct {
	Tiers          []int
	Search         string
	Linear         *int
	SteamIDs       []string
	PrimarySteamID string
	Completion     CompletionFilter
	SortCol        string
	Order          string
	Page           int
	PerPage        int
}

type CompletionFilter string

const (
	CompletionAll        CompletionFilter = "all"
	CompletionCompleted  CompletionFilter = "completed"
	CompletionIncomplete CompletionFilter = "incomplete"
)

type PaginatedMaps struct {
	Maps       []models.Map `json:"maps"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

func GetMaps(database *sql.DB, filters MapFilters) (PaginatedMaps, error) {
	var paginatedMaps PaginatedMaps
	paginatedMaps.Maps = []models.Map{}

	allowedCols := map[string]bool{"comp_per_hour": true, "completions": true, "tier": true, "primary_group": true}
	if !allowedCols[filters.SortCol] {
		filters.SortCol = "comp_per_hour"
	}
	if filters.Order != "asc" && filters.Order != "desc" {
		filters.Order = "desc"
	}

	args := []any{}
	query := buildMapsSelectQuery(filters, &args)
	query += buildMapsOrderByClause(filters, &args)
	query += " LIMIT ? OFFSET ?"
	args = append(args, filters.PerPage)
	args = append(args, (filters.Page-1)*filters.PerPage)

	rows, err := database.Query(query, args...)
	if err != nil {
		return PaginatedMaps{}, err
	}
	defer rows.Close()

	var mapIDs []int
	for rows.Next() {
		var m models.Map
		var playerBestTime sql.NullInt64
		var playerRank sql.NullInt64
		var completed sql.NullInt64
		if err := rows.Scan(
			&m.ID,
			&m.KSFMapID,
			&m.Name,
			&m.Tier,
			&m.Added,
			&m.Completions,
			&m.PlaytimeSeconds,
			&m.CompPerHour,
			&m.Notes,
			&m.Bonus,
			&m.Linear,
			&m.UpdatedAt,
			&playerBestTime,
			&playerRank,
			&completed,
		); err != nil {
			return PaginatedMaps{}, err
		}
		if playerBestTime.Valid {
			value := int(playerBestTime.Int64)
			m.PlayerBestTimeMS = &value
		}
		if playerRank.Valid {
			value := int(playerRank.Int64)
			m.PlayerRank = &value
		}
		if completed.Valid {
			value := completed.Int64 != 0
			m.Completed = &value
		}
		paginatedMaps.Maps = append(paginatedMaps.Maps, m)
		mapIDs = append(mapIDs, m.KSFMapID)
	}

	if err := rows.Err(); err != nil {
		return PaginatedMaps{}, err
	}

	if len(mapIDs) > 0 {
		refreshStates, err := GetMapRecordRefreshStatesForMaps(database, mapIDs)
		if err != nil {
			return PaginatedMaps{}, err
		}
		for i := range paginatedMaps.Maps {
			if state, ok := refreshStates[paginatedMaps.Maps[i].KSFMapID]; ok {
				availableAt := state.CooldownAvailableAt()
				paginatedMaps.Maps[i].RecordRefreshAvailableAt = &availableAt
				paginatedMaps.Maps[i].RecordRefreshStatus = state.Status
				paginatedMaps.Maps[i].RecordRefreshFailedPlayers = state.FailedPlayers
			}
		}

		allPlayers, err := ListPlayers(database)
		if err != nil {
			return PaginatedMaps{}, err
		}
		if len(allPlayers) > 0 {
			allSteamIDs := make([]string, len(allPlayers))
			for i, p := range allPlayers {
				allSteamIDs[i] = p.SteamID
			}
			summariesByMapID, err := GetBestPlayerMapSummariesForMaps(database, allSteamIDs, mapIDs)
			if err != nil {
				return PaginatedMaps{}, err
			}
			for i := range paginatedMaps.Maps {
				paginatedMaps.Maps[i].PlayerRecords = summariesByMapID[paginatedMaps.Maps[i].KSFMapID]
			}
		}
	}

	paginatedMaps.Page = filters.Page
	paginatedMaps.PerPage = filters.PerPage
	paginatedMaps.Total, err = GetMapsCount(database, filters)
	if err != nil {
		return PaginatedMaps{}, err
	}
	paginatedMaps.TotalPages = (paginatedMaps.Total + paginatedMaps.PerPage - 1) / paginatedMaps.PerPage

	return paginatedMaps, nil
}

func buildMapsSelectQuery(filters MapFilters, args *[]any) string {
	query := `
			SELECT 
				id, 
				ksf_map_id, 
				name, 
				tier, 
				added, 
				completions, 
				playtime_seconds,
				comp_per_hour,
				coalesce(notes, ''),
				bonus,
				linear,
				updated_at`

	if filters.PrimarySteamID != "" {
		query += `,
				(
					SELECT pr.surf_time_ms
					FROM player_map_records pr
					WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
					ORDER BY pr.surf_time_ms ASC, pr.id DESC
					LIMIT 1
				) AS player_best_time_ms,
				(
					SELECT pr.rank
					FROM player_map_records pr
					WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
					ORDER BY pr.surf_time_ms ASC, pr.id DESC
					LIMIT 1
				) AS player_rank`
		*args = append(*args, filters.PrimarySteamID, filters.PrimarySteamID)
	} else {
		query += `,
				NULL AS player_best_time_ms,
				NULL AS player_rank`
	}

	if len(filters.SteamIDs) > 0 {
		query += `,
				CASE
					WHEN ` + buildSelectedCompletionCountExpr(filters.SteamIDs, args) + ` = ` + fmt.Sprintf("%d", len(filters.SteamIDs)) + ` THEN 1
					ELSE 0
				END AS completed`
	} else {
		query += `,
				NULL AS completed`
	}

	query += `
			FROM maps
			WHERE 1=1
		`
	query += buildMapsWhereClause(filters, args)

	return query
}

func buildMapsOrderByClause(filters MapFilters, args *[]any) string {
	direction := strings.ToUpper(filters.Order)
	if direction != "ASC" && direction != "DESC" {
		direction = "DESC"
	}

	switch filters.SortCol {
	case "completions":
		return " ORDER BY completions " + direction + ", id ASC"
	case "tier":
		return " ORDER BY tier " + direction + ", id ASC"
	case "primary_group":
		*args = append(*args, filters.PrimarySteamID)
		return `
			ORDER BY COALESCE((
				SELECT pr.group_tier
				FROM player_map_records pr
				WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
				ORDER BY pr.surf_time_ms ASC, pr.id DESC
				LIMIT 1
			), 7) ` + direction + `, comp_per_hour DESC, id ASC`
	default:
		return " ORDER BY comp_per_hour " + direction + ", id ASC"
	}
}

func buildMapsWhereClause(filters MapFilters, args *[]any) string {
	query := ""

	if len(filters.Tiers) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tiers)-1) + "?"
		query += " AND tier IN (" + placeholders + ")"
		for _, tier := range filters.Tiers {
			*args = append(*args, tier)
		}
	}

	if filters.Search != "" {
		query += " AND name LIKE ?"
		*args = append(*args, "%"+filters.Search+"%")
	}

	if filters.Linear != nil {
		query += " AND linear = ?"
		*args = append(*args, *filters.Linear)
	}

	if len(filters.SteamIDs) > 0 {
		switch filters.Completion {
		case CompletionCompleted:
			query += " AND " + buildSelectedCompletionCountExpr(filters.SteamIDs, args) + " = " + fmt.Sprintf("%d", len(filters.SteamIDs))
		case CompletionIncomplete:
			query += " AND " + buildSelectedCompletionCountExpr(filters.SteamIDs, args) + " = 0"
		}
	}

	return query
}

func buildSelectedCompletionCountExpr(steamIDs []string, args *[]any) string {
	placeholders := strings.Repeat("?,", len(steamIDs)-1) + "?"
	for _, steamID := range steamIDs {
		*args = append(*args, steamID)
	}
	return `
		(
			SELECT COUNT(DISTINCT pr.steam_id)
			FROM player_map_records pr
			WHERE pr.ksf_map_id = maps.ksf_map_id
			  AND pr.steam_id IN (` + placeholders + `)
		)
	`
}

func GetMapsCount(database *sql.DB, filters MapFilters) (int, error) {
	args := []any{}
	query := `
			SELECT COUNT(*)
			FROM maps
			WHERE 1=1
		`
	query += buildMapsWhereClause(filters, &args)

	var count int
	err := database.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetBestPlayerMapSummariesForMaps(database *sql.DB, steamIDs []string, mapIDs []int) (map[int][]models.PlayerMapSummary, error) {
	steamPlaceholders := strings.Repeat("?,", len(steamIDs)-1) + "?"
	mapPlaceholders := strings.Repeat("?,", len(mapIDs)-1) + "?"

	args := make([]any, 0, len(steamIDs)+len(mapIDs))
	for _, steamID := range steamIDs {
		args = append(args, steamID)
	}
	for _, mapID := range mapIDs {
		args = append(args, mapID)
	}

	rows, err := database.Query(`
		SELECT steam_id, ksf_map_id, surf_time_ms, rank, group_tier
		FROM player_map_records
		WHERE steam_id IN (`+steamPlaceholders+`)
		  AND ksf_map_id IN (`+mapPlaceholders+`)
		ORDER BY ksf_map_id ASC, steam_id ASC, surf_time_ms ASC, id DESC
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("query best player map summaries: %w", err)
	}
	defer rows.Close()

	type mapKey struct {
		mapID   int
		steamID string
	}

	bestByKey := make(map[mapKey]models.PlayerMapSummary)
	for rows.Next() {
		var steamID string
		var mapID int
		var bestTime int
		var rank sql.NullInt64
		var groupTier sql.NullInt64
		if err := rows.Scan(&steamID, &mapID, &bestTime, &rank, &groupTier); err != nil {
			return nil, fmt.Errorf("scan best player map summary: %w", err)
		}

		key := mapKey{mapID: mapID, steamID: steamID}
		if _, exists := bestByKey[key]; exists {
			continue
		}

		summary := models.PlayerMapSummary{
			SteamID:   steamID,
			Completed: true,
		}
		bestTimeCopy := bestTime
		summary.BestTimeMS = &bestTimeCopy
		if rank.Valid {
			rankValue := int(rank.Int64)
			summary.Rank = &rankValue
		}
		if groupTier.Valid {
			groupTierValue := int(groupTier.Int64)
			summary.GroupTier = &groupTierValue
		}
		bestByKey[key] = summary
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate best player map summaries: %w", err)
	}

	result := make(map[int][]models.PlayerMapSummary, len(mapIDs))
	for _, mapID := range mapIDs {
		summaries := make([]models.PlayerMapSummary, 0, len(steamIDs))
		for _, steamID := range steamIDs {
			key := mapKey{mapID: mapID, steamID: steamID}
			summary, exists := bestByKey[key]
			if !exists {
				summary = models.PlayerMapSummary{
					SteamID:   steamID,
					Completed: false,
				}
			}
			summaries = append(summaries, summary)
		}
		result[mapID] = summaries
	}

	return result, nil
}

func UpsertMaps(database *sql.DB, scrapedMaps []maps.Map) ([]maps.Map, error) {
	newMaps := make([]maps.Map, 0)
	for _, row := range scrapedMaps {
		result, err := database.Exec(`
			INSERT INTO maps (name, ksf_map_id, tier, added, playtime_seconds, bonus, linear)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ksf_map_id) DO NOTHING
		`,
			row.Name,
			row.MapID,
			row.Tier,
			row.Created,
			row.Playtime,
			row.BCount,
			row.IsLinear,
		)

		if err != nil {
			return nil, err
		}

		inserted, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if inserted > 0 {
			newMaps = append(newMaps, row)
			continue
		}

		if _, err := database.Exec(`
			UPDATE maps
			SET name = ?,
				tier = ?,
				added = ?,
				playtime_seconds = ?,
				bonus = ?,
				linear = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE ksf_map_id = ?
		`,
			row.Name,
			row.Tier,
			row.Created,
			row.Playtime,
			row.BCount,
			row.IsLinear,
			row.MapID,
		); err != nil {
			return nil, err
		}
	}

	return newMaps, nil
}

func UpdateMapCompletionsByMapID(database *sql.DB, ksfMapID int, completions int) error {
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	if err := UpdateMapCompletionsByMapIDTx(tx, ksfMapID, completions); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

func UpdateMapCompletionsByMapIDTx(tx *sql.Tx, ksfMapID int, completions int) error {
	_, err := tx.Exec(
		`UPDATE maps
		 SET
            completions = ?,
            comp_per_hour = CASE
            	WHEN playtime_seconds > 0 THEN (? * 3600.0) / playtime_seconds
            	ELSE 0
            END,
            updated_at = CURRENT_TIMESTAMP
         WHERE ksf_map_id = ?`,
		completions, completions, ksfMapID,
	)
	if err != nil {
		return fmt.Errorf("failed to update completions for map_id=%d: %w", ksfMapID, err)
	}
	return nil
}
