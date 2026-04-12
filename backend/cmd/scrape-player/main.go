package main

import (
	"log"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/players"
)

func main() {
	database := db.Open("data/surfstats.db")

	scraper := players.NewKSFScraper()
	mapID, comps, err := scraper.FetchMapCompletionsFromPlayerRecord("STEAM_0:1:75949009", "surf_aircontrol_ksf")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.UpdateMapCompletionsByMapID(database, mapID, comps); err != nil {
		log.Fatal(err)
	}

	println(comps)
}
