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
}

func GetMaps(database *sql.DB, filters MapFilters) ([]models.Map, error) {
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

	rows, err := database.Query(query, args...)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		maps = append(maps, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return maps, nil
}
