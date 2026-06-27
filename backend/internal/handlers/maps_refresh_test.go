package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"surfstats/internal/db"
	"surfstats/internal/jobs"
	"surfstats/migrations"
)

func TestRefreshMapRecordsReturnsAcceptedForNewQueueItem(t *testing.T) {
	database := openHandlerTestDatabase(t)
	availableAt := time.Date(2026, time.May, 21, 12, 0, 0, 0, time.UTC)
	runner := func(_ *sql.DB, ksfMapID int) (jobs.MapRecordRefreshResult, error) {
		if ksfMapID != 10 {
			t.Fatalf("ksfMapID = %d, want 10", ksfMapID)
		}
		return jobs.MapRecordRefreshResult{
			Enqueued:    true,
			Status:      jobs.MapRecordRefreshStatusQueued,
			AvailableAt: &availableAt,
		}, nil
	}

	req := httptest.NewRequest(http.MethodPost, "/api/maps/10/refresh-records", nil)
	rec := httptest.NewRecorder()

	refreshMapRecords(database, runner).ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if !strings.Contains(rec.Body.String(), `"status":"queued"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRefreshMapRecordsReturnsOKForExistingQueueItem(t *testing.T) {
	database := openHandlerTestDatabase(t)
	runner := func(_ *sql.DB, _ int) (jobs.MapRecordRefreshResult, error) {
		return jobs.MapRecordRefreshResult{Status: jobs.MapRecordRefreshStatusRunning}, nil
	}

	req := httptest.NewRequest(http.MethodPost, "/api/maps/10/refresh-records", nil)
	rec := httptest.NewRecorder()

	refreshMapRecords(database, runner).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusOK)
	}
	if strings.TrimSpace(rec.Body.String()) != `{"status":"running"}` {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRefreshMapRecordsReturnsCooldown(t *testing.T) {
	database := openHandlerTestDatabase(t)
	availableAt := time.Date(2026, time.May, 21, 12, 0, 0, 0, time.UTC)
	runner := func(_ *sql.DB, _ int) (jobs.MapRecordRefreshResult, error) {
		return jobs.MapRecordRefreshResult{}, jobs.MapRecordRefreshCooldownError{AvailableAt: availableAt}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/maps/10/refresh-records", nil)
	rec := httptest.NewRecorder()

	refreshMapRecords(database, runner).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
	if !strings.Contains(rec.Body.String(), `"available_at":"2026-05-21T12:00:00Z"`) {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRefreshMapRecordsReturnsNotFound(t *testing.T) {
	database := openHandlerTestDatabase(t)
	runner := func(_ *sql.DB, _ int) (jobs.MapRecordRefreshResult, error) {
		return jobs.MapRecordRefreshResult{}, jobs.MapRecordRefreshNotFoundError{KSFMapID: 10}
	}

	req := httptest.NewRequest(http.MethodPost, "/api/maps/10/refresh-records", nil)
	rec := httptest.NewRecorder()

	refreshMapRecords(database, runner).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestRefreshMapRecordsReturnsMethodNotAllowed(t *testing.T) {
	database := openHandlerTestDatabase(t)
	runner := func(_ *sql.DB, _ int) (jobs.MapRecordRefreshResult, error) {
		return jobs.MapRecordRefreshResult{}, errors.New("should not run")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/maps/10/refresh-records", nil)
	rec := httptest.NewRecorder()

	refreshMapRecords(database, runner).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func openHandlerTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	database := db.Open("file::memory:?cache=shared")
	t.Cleanup(func() { _ = database.Close() })
	if err := db.Migrate(database, migrations.FS); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	return database
}
