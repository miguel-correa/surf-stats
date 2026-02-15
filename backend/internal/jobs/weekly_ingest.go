package jobs

import (
	"database/sql"
	"fmt"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/maps"
	"surfstats/internal/scrapers/players"
)

func RunWeeklyIngestion(database *sql.DB, seedSteamID string) error {
	mapScraper := maps.NewKSFScraper("https://ksf.surf/maps")
	maps, err := mapScraper.FetchMaps("")
	if err != nil {
		return fmt.Errorf("failed to fetch maps: %v", err)
	}

	err = db.UpsertMaps(database, maps)
	if err != nil {
		return fmt.Errorf("error upserting maps: %v", err)
	}

	playerScraper := players.NewKSFScraper()

	for _, m := range maps {
		mapID, comps, err := playerScraper.FetchMapCompletionsFromPlayerRecord(seedSteamID, m.Name)
		if err != nil {
			return fmt.Errorf("error fetching completions for map: %v", err)
		}
		err = db.UpdateMapCompletionsByMapID(database, mapID, comps)
		if err != nil {
			return fmt.Errorf("error updating completions for map: %v", err)
		}
	}

	return nil
}
