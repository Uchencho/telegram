package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/account"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/ws"

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

	defer func() {
		db.Db.Close()
		fmt.Println("Db closed")
	}()
	// db.MigrateDB(db.Db)

	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/", serveHome)
	http.Handle("/api/register", auth.BasicToken(http.HandlerFunc(account.Register)))
	http.Handle("/api/login", auth.BasicToken(http.HandlerFunc(account.Login)))
	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		ws.ChatServer(hub, w, req)
	})

	if err := http.ListenAndServe(defaultServerAddress, nil); err != http.ErrServerClosed {
		log.Println(err)
	}
}
