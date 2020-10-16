package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func databaseURL() string {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found, with error: %v", err)
	}

	dbUrl, present := os.LookupEnv("DATABASE_URL")
	if present {
		return dbUrl
	}
	mysql_conn := os.Getenv("DB_LINK")
	return mysql_conn
}

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

var Db = ConnectDatabase()
