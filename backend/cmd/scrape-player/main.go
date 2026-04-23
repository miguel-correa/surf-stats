package main

import (
	"log"
	"os"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
	"surfstats/migrations"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/scrape-player <steam_id>")
	}

	database := db.Open("data/surfstats.db")
	defer database.Close()

	if err := db.Migrate(database, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	if err := jobs.RunPlayerRecordsIngestion(database, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
