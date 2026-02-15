package db

import (
	"database/sql"
	"fmt"
	"strings"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/maps"
)

type MapFilters struct {
	Tiers   []int
	Search  string
	SortCol string
	Order   string
	Page    int
	PerPage int
}

type PaginatedMaps struct {
	Maps       []models.Map `json:"maps"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PerPage    int          `json:"per_page"`
	TotalPages int          `json:"total_pages"`
}

type PaginatedMapsV2 struct {
	Maps       []models.MapV2 `json:"maps"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

func GetMaps(database *sql.DB, filters MapFilters) (PaginatedMaps, error) {
	var paginatedMaps PaginatedMaps

	args := []any{}
	query := (`
			SELECT id, name, tier, year, completions, hours_played, comp_per_hour, notes
			FROM maps
			WHERE 1=1
		`)

	if len(filters.Tiers) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tiers)-1) + "?"
		query += " AND tier IN (" + placeholders + ")"
		for _, tier := range filters.Tiers {
			args = append(args, tier)
		}
	}

	if filters.Search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filters.Search+"%")
	}

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
			return PaginatedMaps{}, err
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

func GetMapsV2(database *sql.DB, filters MapFilters) (PaginatedMapsV2, error) {
	var paginatedMaps PaginatedMapsV2

	args := []any{}
	query := (`
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
				updated_at
			FROM maps
			WHERE 1=1
		`)

	if len(filters.Tiers) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tiers)-1) + "?"
		query += " AND tier IN (" + placeholders + ")"
		for _, tier := range filters.Tiers {
			args = append(args, tier)
		}
	}

	if filters.Search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filters.Search+"%")
	}

	query += " ORDER BY " + filters.SortCol + " " + filters.Order
	query += " LIMIT ? OFFSET ?"
	args = append(args, filters.PerPage)
	args = append(args, (filters.Page-1)*filters.PerPage)

	rows, err := database.Query(query, args...)
	if err != nil {
		return PaginatedMapsV2{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var m models.MapV2
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
		); err != nil {
			return PaginatedMapsV2{}, err
		}
		paginatedMaps.Maps = append(paginatedMaps.Maps, m)
	}

	if err := rows.Err(); err != nil {
		return PaginatedMapsV2{}, err
	}

	paginatedMaps.Page = filters.Page
	paginatedMaps.PerPage = filters.PerPage
	paginatedMaps.Total, err = GetMapsCountV2(database, filters)
	if err != nil {
		return PaginatedMapsV2{}, err
	}
	paginatedMaps.TotalPages = (paginatedMaps.Total + paginatedMaps.PerPage - 1) / paginatedMaps.PerPage
	return paginatedMaps, nil
}

func GetMapsCount(database *sql.DB, filters MapFilters) (int, error) {
	args := []any{}
	query := (`
			SELECT COUNT(*)
			FROM maps
			WHERE 1=1
		`)

	if len(filters.Tiers) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tiers)-1) + "?"
		query += " AND tier IN (" + placeholders + ")"
		for _, tier := range filters.Tiers {
			args = append(args, tier)
		}
	}

	if filters.Search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filters.Search+"%")
	}

	var count int
	err := database.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetMapsCountV2(database *sql.DB, filters MapFilters) (int, error) {
	args := []any{}
	query := (`
			SELECT COUNT(*)
			FROM maps
			WHERE 1=1
		`)

	if len(filters.Tiers) > 0 {
		placeholders := strings.Repeat("?,", len(filters.Tiers)-1) + "?"
		query += " AND tier IN (" + placeholders + ")"
		for _, tier := range filters.Tiers {
			args = append(args, tier)
		}
	}

	if filters.Search != "" {
		query += " AND name LIKE ?"
		args = append(args, "%"+filters.Search+"%")
	}

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
