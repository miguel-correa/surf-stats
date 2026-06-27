package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/models"
	"surfstats/internal/scrapers/players"
)

const (
	MapRecordRefreshStatusQueued              = "queued"
	MapRecordRefreshStatusRunning             = "running"
	MapRecordRefreshStatusCompleted           = "completed"
	MapRecordRefreshStatusCompletedWithErrors = "completed_with_errors"
)

type MapRecordRefreshResult struct {
	Enqueued         bool       `json:"-"`
	Status           string     `json:"status"`
	AvailableAt      *time.Time `json:"available_at,omitempty"`
	RefreshedPlayers int        `json:"refreshed_players"`
	FailedPlayers    int        `json:"failed_players"`
	Failures         []string   `json:"failures,omitempty"`
}

type MapRecordRefreshCooldownError struct {
	AvailableAt time.Time
}

func (e MapRecordRefreshCooldownError) Error() string {
	return fmt.Sprintf("map record refresh is on cooldown until %s", e.AvailableAt.Format(time.RFC3339))
}

type MapRecordRefreshNotFoundError struct {
	KSFMapID int
}

func (e MapRecordRefreshNotFoundError) Error() string {
	return fmt.Sprintf("map not found map_id=%d", e.KSFMapID)
}

func EnqueueMapRecordRefresh(database *sql.DB, ksfMapID int) (MapRecordRefreshResult, error) {
	return enqueueMapRecordRefresh(database, ksfMapID, time.Now)
}

func enqueueMapRecordRefresh(database *sql.DB, ksfMapID int, now func() time.Time) (MapRecordRefreshResult, error) {
	mapRef, err := db.GetMapRefByKSFMapID(database, ksfMapID)
	if err != nil {
		return MapRecordRefreshResult{}, err
	}
	if mapRef == nil {
		return MapRecordRefreshResult{}, MapRecordRefreshNotFoundError{KSFMapID: ksfMapID}
	}

	enqueue, err := db.EnqueueMapRecordRefresh(database, ksfMapID, now())
	if err != nil {
		return MapRecordRefreshResult{}, err
	}
	if enqueue.Cooldown {
		return MapRecordRefreshResult{}, MapRecordRefreshCooldownError{AvailableAt: enqueue.AvailableAt}
	}

	result := resultFromRefreshState(enqueue.State, enqueue.AvailableAt)
	result.Enqueued = enqueue.Enqueued
	return result, nil
}

func RunMapRecordRefreshQueueWorker(ctx context.Context, database *sql.DB, delay time.Duration) {
	RunMapRecordRefreshQueueWorkerWithScraper(ctx, database, delay, players.NewKSFScraper())
}

func RunMapRecordRefreshQueueWorkerWithScraper(ctx context.Context, database *sql.DB, delay time.Duration, scraper mainRecordScraper) {
	for {
		processed, err := ProcessNextMapRecordRefresh(ctx, database, scraper, time.Now)
		if err != nil {
			log.Printf("map-record-refresh-worker: %v", err)
		}

		wait := 2 * time.Second
		if processed {
			wait = delay
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func ProcessNextMapRecordRefresh(ctx context.Context, database *sql.DB, scraper mainRecordScraper, now func() time.Time) (bool, error) {
	mapRef, err := db.ClaimNextQueuedMapRecordRefresh(database, now())
	if err != nil {
		return false, err
	}
	if mapRef == nil {
		return false, nil
	}

	_, err = runClaimedMapRecordRefresh(ctx, database, *mapRef, scraper, now)
	return true, err
}

func runClaimedMapRecordRefresh(ctx context.Context, database *sql.DB, mapRef db.MapRef, scraper mainRecordScraper, now func() time.Time) (MapRecordRefreshResult, error) {
	result := MapRecordRefreshResult{Status: MapRecordRefreshStatusCompleted}
	status := MapRecordRefreshStatusCompleted
	var failures []string

	storedPlayers, err := db.ListPlayers(database)
	if err != nil {
		_ = db.FinishMapRecordRefresh(database, mapRef.KSFMapID, MapRecordRefreshStatusCompletedWithErrors, 0, 1, err.Error(), now())
		return MapRecordRefreshResult{}, err
	}

	for _, player := range storedPlayers {
		select {
		case <-ctx.Done():
			failures = append(failures, fmt.Sprintf("worker stopped: %v", ctx.Err()))
			result.FailedPlayers = len(failures)
			result.Failures = failures
			_ = db.FinishMapRecordRefresh(database, mapRef.KSFMapID, MapRecordRefreshStatusCompletedWithErrors, result.RefreshedPlayers, result.FailedPlayers, strings.Join(failures, "\n"), now())
			return result, ctx.Err()
		default:
		}

		record, err := fetchMainMapRecordWithRetry(scraper, player.SteamID, mapRef.Name, mapRef.KSFMapID)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", player.SteamID, err))
			continue
		}
		if record == nil {
			result.RefreshedPlayers++
			continue
		}

		record.KSFMapID = mapRef.KSFMapID
		if record.SteamID == "" {
			record.SteamID = player.SteamID
		}
		if record.PlayerName == "" {
			record.PlayerName = player.Name
		}
		if err := upsertPlayerFromRecord(database, player, *record); err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", player.SteamID, err))
			continue
		}
		inserted, err := db.SavePlayerMapRecordIfImproved(database, *record)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", player.SteamID, err))
			continue
		}
		if inserted {
			log.Printf(
				"map-record-refresh: new record steam_id=%s map=%s time_ms=%d rank=%d",
				record.SteamID,
				mapRef.Name,
				record.SurfTimeMS,
				record.Rank,
			)
		}
		result.RefreshedPlayers++
	}

	result.FailedPlayers = len(failures)
	result.Failures = failures
	if len(failures) > 0 {
		status = MapRecordRefreshStatusCompletedWithErrors
	}
	result.Status = status

	lastError := ""
	if len(failures) > 0 {
		lastError = strings.Join(failures, "\n")
	}
	if err := db.FinishMapRecordRefresh(database, mapRef.KSFMapID, status, result.RefreshedPlayers, result.FailedPlayers, lastError, now()); err != nil {
		return result, err
	}

	return result, nil
}

func upsertPlayerFromRecord(database *sql.DB, fallback models.Player, record models.PlayerMapRecord) error {
	playerID := record.PlayerID
	player := models.Player{
		SteamID:  record.SteamID,
		PlayerID: &playerID,
		Name:     record.PlayerName,
	}
	if player.SteamID == "" {
		player.SteamID = fallback.SteamID
	}
	if player.Name == "" {
		player.Name = fallback.Name
	}
	if player.Name == "" {
		player.Name = player.SteamID
	}
	if playerID == 0 && fallback.PlayerID != nil {
		player.PlayerID = fallback.PlayerID
	}
	return db.UpsertPlayer(database, player)
}

func resultFromRefreshState(state *db.MapRecordRefreshState, availableAt time.Time) MapRecordRefreshResult {
	if state == nil {
		return MapRecordRefreshResult{}
	}
	availableAtCopy := availableAt
	if availableAtCopy.IsZero() {
		availableAtCopy = state.CooldownAvailableAt()
	}
	return MapRecordRefreshResult{
		Status:           state.Status,
		AvailableAt:      &availableAtCopy,
		RefreshedPlayers: state.RefreshedPlayers,
		FailedPlayers:    state.FailedPlayers,
	}
}

func IsMapRecordRefreshCooldown(err error) (MapRecordRefreshCooldownError, bool) {
	var cooldownErr MapRecordRefreshCooldownError
	ok := errors.As(err, &cooldownErr)
	return cooldownErr, ok
}

func IsMapRecordRefreshNotFound(err error) bool {
	var notFoundErr MapRecordRefreshNotFoundError
	return errors.As(err, &notFoundErr)
}
