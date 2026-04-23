package main

import (
	"log"
	"surfstats/internal/db"
	"surfstats/internal/scrapers/maps"
	"surfstats/migrations"
)

func main() {

	database := db.Open("data/surfstats.db")
	defer database.Close()

	if err := db.Migrate(database, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}

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
