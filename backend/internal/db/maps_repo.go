package db

import (
	"database/sql"
	"surfstats/internal/models"
)

type MapFilters struct {
	Tier    *int
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

	if filters.Tier != nil {
		query += " AND tier = ?"
		args = append(args, *filters.Tier)
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
