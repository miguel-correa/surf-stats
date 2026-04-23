package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"surfstats/internal/db"
)

func GetPlayers(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		players, err := db.ListPlayers(database)
		if err != nil {
			log.Printf("GetPlayers error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(players)
	}
}
