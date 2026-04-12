package main

import (
	"log"
	"os"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/scrape-player <steam_id>")
	}

	database := db.Open("data/surfstats.db")
	defer database.Close()

	if err := jobs.RunPlayerRecordsIngestion(database, os.Args[1]); err != nil {
		log.Fatal(err)
	}
}
