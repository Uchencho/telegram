package db

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"

	// Needed
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateDB helps apply migrations to a database
func MigrateDB(db *sql.DB, driver database.Driver, dbType string) {

	currentDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	migDir := strings.Split(currentDir, "telegram")[0] + "telegram/db/migration/"

	var migrationDir = flag.String("migration files", migDir, "Directory where the migration file exists")
	flag.Parse()

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationDir), dbType, driver,
	)
	if err != nil {
		log.Fatalf("Error encountered in creating new db instance, %v", err)
	}
	err = m.Up()

	if err != nil && err == migrate.ErrNoChange {
		log.Println("No new migration file")
		return
	} else if err != nil {
		log.Fatalf("Error in migrating with error, %v", err)
	}
	fmt.Println("Migrated successfully")
}
