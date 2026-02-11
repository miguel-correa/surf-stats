package main

import (
	"log"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/maps"
)

func main() {

	database := db.Open("data/surfstats.db")

	scraper := maps.NewKSFScraper("https://ksf.surf/maps")
	scrapedMaps, err := scraper.FetchMaps("data/ksf_maps.html")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.UpsertMaps(database, scrapedMaps); err != nil {
		log.Fatal(err)
	}

	// os.WriteFile("data/ksf_maps.html", html, 0644)
}
