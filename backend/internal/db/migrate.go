package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"sort"
)

func Migrate(database *sql.DB, migrationsFS fs.FS) error {
	if _, err := database.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY
	)`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	entries, err := fs.Glob(migrationsFS, "*.sql")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	sort.Strings(entries)

	for _, name := range entries {
		var count int
		if err := database.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE filename = ?", name).Scan(&count); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := database.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}
		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("execute migration %s: %w", name, err)
		}
		if _, err := tx.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", name); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
		log.Printf("applied migration: %s", name)
	}

	return nil
}
