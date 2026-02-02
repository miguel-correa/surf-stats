package db

import (
	"database/sql"
	"strings"
	"surfstats/internal/models"
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
