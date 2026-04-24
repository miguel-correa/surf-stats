package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"surfstats/internal/db"
	"surfstats/internal/jobs"
	"surfstats/migrations"
)

const failedMapsFile = "data/failed-maps.txt"

func main() {
	retryFailed := flag.Bool("retry-failed", false, "retry only maps listed in "+failedMapsFile)
	flag.Parse()

	database := db.Open("data/surfstats.db")
	defer database.Close()

	if err := db.Migrate(database, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	steamID := "STEAM_0:1:75949009"

	var failedMaps []string
	var err error

	if *retryFailed {
		mapNames, readErr := readFailedMaps()
		if readErr != nil {
			log.Fatalf("failed to read %s: %v", failedMapsFile, readErr)
		}
		if len(mapNames) == 0 {
			fmt.Println("no failed maps to retry")
			return
		}
		fmt.Printf("retrying %d failed map(s)\n", len(mapNames))
		failedMaps, err = jobs.RunCompletionsForMaps(database, steamID, mapNames)
	} else {
		failedMaps, err = jobs.RunWeeklyIngestion(database, steamID)
	}

	if err != nil {
		fmt.Printf("error running weekly ingestion job: %v\n", err)
		return
	}

	if len(failedMaps) > 0 {
		if writeErr := os.WriteFile(failedMapsFile, []byte(strings.Join(failedMaps, "\n")+"\n"), 0644); writeErr != nil {
			log.Printf("warning: could not write %s: %v", failedMapsFile, writeErr)
		} else {
			fmt.Printf("wrote %d failed map(s) to %s — re-run with --retry-failed\n", len(failedMaps), failedMapsFile)
		}
	} else {
		os.Remove(failedMapsFile)
		fmt.Println("finished maps weekly ingestion job")
	}
}

func readFailedMaps() ([]string, error) {
	data, err := os.ReadFile(failedMapsFile)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if name := strings.TrimSpace(line); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
