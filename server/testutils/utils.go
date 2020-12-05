package testutils

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"

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
		log.Fatal("Failed to open DB with err", err)
	}
	log.Println("DB Open")
	return db
}

// GetTestDriver returns a mysql testdriver for migration
func GetTestDriver(db *sql.DB) database.Driver {
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		log.Fatal("Could not get sqlite2 driver for test", err)
	}
	return driver
}

// DropDB returns the functionality to drop the DB
func DropDB() {
	log.Println("Called dropDB")
	err := os.Remove("test_database.db")
	if err != nil {
		log.Println("File does not exist")
	}
}
