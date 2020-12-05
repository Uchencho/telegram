package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Indirect import needed
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// RetrieveUserDetailsFunc
// type RetrieveUserDetailsFunc func(db *sql.DB, email string) (auth.User, error)

func databaseURL() string {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found, with error: %v", err)
	}

	dbURL, present := os.LookupEnv("DATABASE_URL")
	if present {
		return dbURL
	}
	mySQLConn := os.Getenv("DB_LINK")
	return mySQLConn
}

// ConnectDatabase connects to a database
func ConnectDatabase() *sql.DB {
	db, err := sql.Open("mysql", databaseURL())
	if err != nil {
		log.Fatalf("Failed to connect to database with error, %s", err)
	}

	dbErr := db.Ping()
	if dbErr != nil {
		log.Fatalf("Failed to ping database with error, %s", dbErr)
	}
	fmt.Println("Connected successfully")
	return db
}

// Db is exported to give the functionality of interacting with the database
var Db = ConnectDatabase()
