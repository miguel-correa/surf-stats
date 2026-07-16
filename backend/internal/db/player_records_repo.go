package db

import (
	"database/sql"
	"fmt"
	"time"

	"surfstats/internal/models"
)

type MapRef struct {
	KSFMapID int
	Name     string
}

func ListMapsForPlayerRecordIngestion(database *sql.DB) ([]MapRef, error) {
	rows, err := database.Query(`SELECT ksf_map_id, name FROM maps ORDER BY id ASC`)
	if err != nil {
		return nil, fmt.Errorf("query maps for player record ingestion: %w", err)
	}
	defer rows.Close()

	var maps []MapRef
	for rows.Next() {
		var m MapRef
		if err := rows.Scan(&m.KSFMapID, &m.Name); err != nil {
			return nil, fmt.Errorf("scan map for player record ingestion: %w", err)
		}
		maps = append(maps, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate maps for player record ingestion: %w", err)
	}

	return maps, nil
}

func GetLatestPlayerMapRecord(database *sql.DB, steamID string, ksfMapID int) (*models.PlayerMapRecord, error) {
	row := database.QueryRow(`
		SELECT
			id,
			steam_id,
			player_id,
			ksf_map_id,
			surf_time_ms,
			rank,
			total_ranks,
			date_set,
			completions,
			group_tier,
			scraped_at
		FROM player_map_records
		WHERE steam_id = ? AND ksf_map_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, steamID, ksfMapID)

	var record models.PlayerMapRecord
	var dateSet sql.NullTime
	var completions sql.NullInt64
	var groupTier sql.NullInt64

	err := row.Scan(
		&record.ID,
		&record.SteamID,
		&record.PlayerID,
		&record.KSFMapID,
		&record.SurfTimeMS,
		&record.Rank,
		&record.TotalRanks,
		&dateSet,
		&completions,
		&groupTier,
		&record.ScrapedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest player map record steam_id=%s map_id=%d: %w", steamID, ksfMapID, err)
	}

	if dateSet.Valid {
		record.DateSet = &dateSet.Time
	}
	if completions.Valid {
		value := int(completions.Int64)
		record.Completions = &value
	}
	if groupTier.Valid {
		value := int(groupTier.Int64)
		record.GroupTier = &value
	}

	return &record, nil
}

func SavePlayerMapRecordIfImproved(database *sql.DB, record models.PlayerMapRecord) (bool, error) {
	latest, err := GetLatestPlayerMapRecord(database, record.SteamID, record.KSFMapID)
	if err != nil {
		return false, err
	}

	if latest != nil {
		switch {
		case record.SurfTimeMS == latest.SurfTimeMS:
			if !playerMapRecordMetadataChanged(record, *latest) {
				return false, nil
			}
		case record.SurfTimeMS > latest.SurfTimeMS:
			return false, nil
		}
	}

	if err := InsertPlayerMapRecord(database, record); err != nil {
		return false, err
	}
	return true, nil
}

func playerMapRecordMetadataChanged(next models.PlayerMapRecord, latest models.PlayerMapRecord) bool {
	if next.Rank != latest.Rank ||
		next.TotalRanks != latest.TotalRanks {
		return true
	}

	if !sameOptionalTime(next.DateSet, latest.DateSet) {
		return true
	}
	if !sameOptionalInt(next.Completions, latest.Completions) {
		return true
	}
	return !sameOptionalInt(next.GroupTier, latest.GroupTier)
}

func sameOptionalInt(a *int, b *int) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func sameOptionalTime(a *time.Time, b *time.Time) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return a.Equal(*b)
	}
}

func InsertPlayerMapRecord(database *sql.DB, record models.PlayerMapRecord) error {
	_, err := database.Exec(`
		INSERT INTO player_map_records (
			steam_id,
			player_id,
			ksf_map_id,
			surf_time_ms,
			rank,
			total_ranks,
			date_set,
			completions,
			group_tier
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.SteamID,
		record.PlayerID,
		record.KSFMapID,
		record.SurfTimeMS,
		record.Rank,
		record.TotalRanks,
		record.DateSet,
		record.Completions,
		record.GroupTier,
	)
	if err != nil {
		return fmt.Errorf("insert player map record steam_id=%s map_id=%d: %w", record.SteamID, record.KSFMapID, err)
	}
	return nil
}
