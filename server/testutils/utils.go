package testutils

import (
	"database/sql"
	"log"
	"os"

	// Sqlite3
	_ "github.com/mattn/go-sqlite3"
)

// CreateDB returns an sqlite db for testing
func CreateDB() *sql.DB {
	_, err := os.Create("test_database.db")
	if err != nil {
		log.Println("Error in creating test db")
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "test_database.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// DropDB returns the functionality to drop the DB
func DropDB() {
	err := os.Remove("test_database.db")
	if err != nil {
		log.Println("File does not exist")
	}
}
