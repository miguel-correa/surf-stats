package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

func Open(path string) *sql.DB {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
