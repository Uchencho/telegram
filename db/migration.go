package db

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateDB(db *sql.DB) {

	var migrationDir = flag.String("migration files", "./db/migration/",
		"Directory where the migration file exists")
	flag.Parse()

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to connect with error %s", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", *migrationDir), "mysql", driver,
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
