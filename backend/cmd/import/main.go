package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"os"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run cmd/import/main.go <file.csv>")
	}

	filePath := os.Args[1]

	db, err := sql.Open("sqlite", "surfstats.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for i, row := range records {
		// Skip header
		if i == 0 {
			continue
		}

		name := strings.TrimSpace(row[0])
		tier := parseInt(row[1])
		year := parseInt(row[2])
		completions := parseInt(row[3])
		hours := parseInt(row[4])
		notes := ""
		if len(row) > 6 {
			notes = row[6]
		}

		compPerHours := 0.0
		if hours > 0 {
			compPerHours = float64(completions) / float64(hours)
		}

		_, err := db.Exec(`
			INSERT INTO maps (name, tier, year, completions, hours_played, comp_per_hour, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
			  tier = excluded.tier,
			  year = excluded.year,
			  completions = excluded.completions,
			  hours_played = excluded.hours_played,
			  comp_per_hour = excluded.comp_per_hour,
			  notes = excluded.notes,
			  updated_at = CURRENT_TIMESTAMP
		`,
			name,
			tier,
			year,
			completions,
			hours,
			compPerHours,
			notes,
		)

		if err != nil {
			log.Fatalf("error importing %s: %v", name, err)
		}
	}

	log.Println("CSV import completed successfully")
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(strings.TrimSpace(s))
	return i
}
