package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/app"

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

func main() {

	mySQL := db.ConnectDatabase()

	defer func() {
		mySQL.Close()
		fmt.Println("Db closed")
	}()
	db.MigrateDB(mySQL)

	a := app.NewApp(mySQL)

	log.Println("Running on address: ", defaultServerAddress)
	if err := http.ListenAndServe(defaultServerAddress, a.Handler()); err != http.ErrServerClosed {
		log.Println(err)
	}
}
