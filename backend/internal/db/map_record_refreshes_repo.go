package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

const mapRecordRefreshCooldown = time.Minute

type MapRecordRefreshState struct {
	KSFMapID         int
	QueuedAt         *time.Time
	LastStarted      time.Time
	LastFinished     *time.Time
	Status           string
	RefreshedPlayers int
	FailedPlayers    int
	LastError        string
}

type MapRecordRefreshEnqueue struct {
	Enqueued    bool
	Existing    bool
	Cooldown    bool
	AvailableAt time.Time
	State       *MapRecordRefreshState
}

func EnqueueMapRecordRefresh(database *sql.DB, ksfMapID int, now time.Time) (MapRecordRefreshEnqueue, error) {
	now = now.UTC().Truncate(time.Second)

	state, err := GetMapRecordRefreshState(database, ksfMapID)
	if err != nil {
		return MapRecordRefreshEnqueue{}, err
	}
	if state != nil {
		switch state.Status {
		case "queued", "running":
			return MapRecordRefreshEnqueue{
				Existing:    true,
				AvailableAt: state.CooldownAvailableAt(),
				State:       state,
			}, nil
		}

		availableAt := state.CooldownAvailableAt()
		if availableAt.After(now) {
			return MapRecordRefreshEnqueue{
				Cooldown:    true,
				AvailableAt: availableAt,
				State:       state,
			}, nil
		}
	}

	_, err = database.Exec(`
		INSERT INTO map_record_refreshes (
			ksf_map_id,
			queued_at,
			last_started_at,
			last_finished_at,
			status,
			refreshed_players,
			failed_players,
			last_error
		) VALUES (?, ?, ?, NULL, 'queued', 0, 0, NULL)
		ON CONFLICT(ksf_map_id) DO UPDATE SET
			queued_at = excluded.queued_at,
			last_started_at = excluded.last_started_at,
			last_finished_at = NULL,
			status = 'queued',
			refreshed_players = 0,
			failed_players = 0,
			last_error = NULL
	`, ksfMapID, formatSQLiteTime(now), formatSQLiteTime(now))
	if err != nil {
		return MapRecordRefreshEnqueue{}, fmt.Errorf("enqueue map record refresh map_id=%d: %w", ksfMapID, err)
	}

	state, err = GetMapRecordRefreshState(database, ksfMapID)
	if err != nil {
		return MapRecordRefreshEnqueue{}, err
	}

	return MapRecordRefreshEnqueue{
		Enqueued:    true,
		AvailableAt: now.Add(mapRecordRefreshCooldown),
		State:       state,
	}, nil
}

func ClaimNextQueuedMapRecordRefresh(database *sql.DB, now time.Time) (*MapRef, error) {
	now = now.UTC().Truncate(time.Second)
	tx, err := database.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin claim next map record refresh: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow(`
		SELECT r.ksf_map_id, m.name
		FROM map_record_refreshes r
		JOIN maps m ON m.ksf_map_id = r.ksf_map_id
		WHERE r.status = 'queued'
		ORDER BY r.queued_at ASC, r.last_started_at ASC
		LIMIT 1
	`)

	var m MapRef
	if err := row.Scan(&m.KSFMapID, &m.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select queued map record refresh: %w", err)
	}

	result, err := tx.Exec(`
		UPDATE map_record_refreshes
		SET status = 'running',
			last_started_at = ?,
			last_finished_at = NULL
		WHERE ksf_map_id = ? AND status = 'queued'
	`, formatSQLiteTime(now), m.KSFMapID)
	if err != nil {
		return nil, fmt.Errorf("claim queued map record refresh map_id=%d: %w", m.KSFMapID, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read claim queued rows affected map_id=%d: %w", m.KSFMapID, err)
	}
	if rowsAffected == 0 {
		return nil, nil
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim queued map record refresh map_id=%d: %w", m.KSFMapID, err)
	}
	return &m, nil
}

func FinishMapRecordRefresh(database *sql.DB, ksfMapID int, status string, refreshedPlayers int, failedPlayers int, lastError string, now time.Time) error {
	now = now.UTC().Truncate(time.Second)
	_, err := database.Exec(`
		UPDATE map_record_refreshes
		SET last_finished_at = ?,
			status = ?,
			refreshed_players = ?,
			failed_players = ?,
			last_error = ?
		WHERE ksf_map_id = ?
	`, formatSQLiteTime(now), status, refreshedPlayers, failedPlayers, nullString(lastError), ksfMapID)
	if err != nil {
		return fmt.Errorf("finish map record refresh map_id=%d: %w", ksfMapID, err)
	}
	return nil
}

func GetMapRecordRefreshState(database *sql.DB, ksfMapID int) (*MapRecordRefreshState, error) {
	rows, err := GetMapRecordRefreshStatesForMaps(database, []int{ksfMapID})
	if err != nil {
		return nil, err
	}
	state, ok := rows[ksfMapID]
	if !ok {
		return nil, nil
	}
	return &state, nil
}

func GetMapRecordRefreshStatesForMaps(database *sql.DB, mapIDs []int) (map[int]MapRecordRefreshState, error) {
	result := make(map[int]MapRecordRefreshState, len(mapIDs))
	if len(mapIDs) == 0 {
		return result, nil
	}

	placeholders := strings.Repeat("?,", len(mapIDs)-1) + "?"
	args := make([]any, 0, len(mapIDs))
	for _, mapID := range mapIDs {
		args = append(args, mapID)
	}

	rows, err := database.Query(`
		SELECT
			ksf_map_id,
			queued_at,
			last_started_at,
			last_finished_at,
			status,
			refreshed_players,
			failed_players,
			coalesce(last_error, '')
		FROM map_record_refreshes
		WHERE ksf_map_id IN (`+placeholders+`)
	`, args...)
	if err != nil {
		return nil, fmt.Errorf("query map record refresh states: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var state MapRecordRefreshState
		var queuedAt sql.NullTime
		var lastFinished sql.NullTime
		if err := rows.Scan(
			&state.KSFMapID,
			&queuedAt,
			&state.LastStarted,
			&lastFinished,
			&state.Status,
			&state.RefreshedPlayers,
			&state.FailedPlayers,
			&state.LastError,
		); err != nil {
			return nil, fmt.Errorf("scan map record refresh state: %w", err)
		}
		if queuedAt.Valid {
			state.QueuedAt = &queuedAt.Time
		}
		if lastFinished.Valid {
			state.LastFinished = &lastFinished.Time
		}
		result[state.KSFMapID] = state
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate map record refresh states: %w", err)
	}

	return result, nil
}

func GetMapRefByKSFMapID(database *sql.DB, ksfMapID int) (*MapRef, error) {
	row := database.QueryRow(`SELECT ksf_map_id, name FROM maps WHERE ksf_map_id = ?`, ksfMapID)

	var m MapRef
	if err := row.Scan(&m.KSFMapID, &m.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get map ref map_id=%d: %w", ksfMapID, err)
	}
	return &m, nil
}

func (s MapRecordRefreshState) CooldownAvailableAt() time.Time {
	if s.QueuedAt != nil {
		return s.QueuedAt.Add(mapRecordRefreshCooldown)
	}
	return s.LastStarted.Add(mapRecordRefreshCooldown)
}

func formatSQLiteTime(t time.Time) string {
	return t.UTC().Format("2006-01-02 15:04:05")
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
