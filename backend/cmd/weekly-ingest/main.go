package main

import (
	"fmt"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
)

func main() {
	database := db.Open("data/surfstats.db")
	defer database.Close()
	err := jobs.RunWeeklyIngestion(database, "STEAM_0:1:75949009")
	if err != nil {
		fmt.Printf("error running weekly ingestion job: %v\n", err)
		return
	}
	fmt.Println("finished maps weekly ingestion job")
}
