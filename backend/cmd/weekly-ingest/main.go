package main

import (
	"fmt"
	"log"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
	"surfstats/migrations"
)

func main() {
	database := db.Open("data/surfstats.db")
	defer database.Close()

	if err := db.Migrate(database, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	err := jobs.RunWeeklyIngestion(database, "STEAM_0:1:75949009")
	if err != nil {
		fmt.Printf("error running weekly ingestion job: %v\n", err)
		return
	}
	fmt.Println("finished maps weekly ingestion job")
}
