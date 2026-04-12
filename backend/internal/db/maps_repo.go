package db

import (
	"database/sql"
	"fmt"
	"strings"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/maps"
)

type MapFilters struct {
	Tiers      []int
	Search     string
	Linear     *int
	SteamID    string
	Completion CompletionFilter
	SortCol    string
	Order      string
	Page       int
	PerPage    int
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

	args := []any{}
	query := buildMapsSelectQuery(filters, &args)
	query += " ORDER BY " + filters.SortCol + " " + filters.Order
	query += " LIMIT ? OFFSET ?"
	args = append(args, filters.PerPage)
	args = append(args, (filters.Page-1)*filters.PerPage)

	rows, err := database.Query(query, args...)
	if err != nil {
		return PaginatedMaps{}, err
	}
	defer rows.Close()

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
	}

	if err := rows.Err(); err != nil {
		return PaginatedMaps{}, err
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

	if filters.SteamID != "" {
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
				) AS player_rank,
				CASE
					WHEN EXISTS (
						SELECT 1
						FROM player_map_records pr
						WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
					) THEN 1
					ELSE 0
				END AS completed`
		*args = append(*args, filters.SteamID, filters.SteamID, filters.SteamID)
	} else {
		query += `,
				NULL AS player_best_time_ms,
				NULL AS player_rank,
				NULL AS completed`
	}

	query += `
			FROM maps
			WHERE 1=1
		`
	query += buildMapsWhereClause(filters, args)

	return query
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

	if filters.SteamID != "" {
		switch filters.Completion {
		case CompletionCompleted:
			query += ` AND EXISTS (
				SELECT 1
				FROM player_map_records pr
				WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
			)`
			*args = append(*args, filters.SteamID)
		case CompletionIncomplete:
			query += ` AND NOT EXISTS (
				SELECT 1
				FROM player_map_records pr
				WHERE pr.steam_id = ? AND pr.ksf_map_id = maps.ksf_map_id
			)`
			*args = append(*args, filters.SteamID)
		}
	}

	return query
}

func GetMapsCount(database *sql.DB, filters MapFilters) (int, error) {
	args := []any{}
	query := (`
			SELECT COUNT(*)
			FROM maps
			WHERE 1=1
		`)
	query += buildMapsWhereClause(filters, &args)
	var count int
	err := database.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func UpsertMaps(database *sql.DB, maps []maps.Map) error {

	for _, row := range maps {
		_, err := database.Exec(`
			INSERT INTO maps (name, ksf_map_id, tier, added, playtime_seconds, bonus, linear)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ksf_map_id) DO UPDATE SET
			  name = excluded.name,
			  tier = excluded.tier,
			  added = excluded.added,
			  playtime_seconds = excluded.playtime_seconds,
			  bonus = excluded.bonus,
			  linear = excluded.linear,
			  updated_at = CURRENT_TIMESTAMP
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
			return err
		}
	}

	return nil
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
