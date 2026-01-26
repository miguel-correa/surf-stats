package main

import (
	"log"
	"net/http"
	"surfstats/internal/db"
	"surfstats/internal/handlers"
)

func main() {
	database := db.Open("surfstats.db")

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.Handle("/api/maps", handlers.GetMaps(database))

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
