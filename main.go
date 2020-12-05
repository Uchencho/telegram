package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/app"

	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/joho/godotenv"
)

const (
	// default server address
	defaultServerAddress = "127.0.0.1:8000"
)

func serveHome(w http.ResponseWriter, req *http.Request) {
	log.Println(req.URL)

	if req.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if req.Method != http.MethodGet {
		http.Error(w, "Method Not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, req, "home.html")
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("No .env file found, with error: %s", err)
	}
}

func inititeMYSQL() *sql.DB {
	mySQL := db.ConnectDatabase()
	driver, err := mysql.WithInstance(mySQL, &mysql.Config{})
	if err != nil {
		log.Fatalf("Failed to connect with error %s", err)
	}

	currentDB := os.Getenv("DB_NAME")
	db.MigrateDB(mySQL, driver, currentDB)
	return mySQL
}

func main() {

	mySQL := inititeMYSQL()
	defer func() {
		mySQL.Close()
		fmt.Println("Db closed")
	}()

	a := app.NewApp(mySQL)

	log.Println("Running on address: ", defaultServerAddress)
	if err := http.ListenAndServe(defaultServerAddress, a.Handler()); err != http.ErrServerClosed {
		log.Println(err)
	}
}
