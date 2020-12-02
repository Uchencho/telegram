package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/account"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/utils"
	"github.com/Uchencho/telegram/server/ws"
	"github.com/gorilla/mux"

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
	db.MigrateDB(db.Db)

	router := mux.NewRouter()
	router.NotFoundHandler = auth.BasicToken(http.HandlerFunc(utils.NotAvailabe))

	router.HandleFunc("/", serveHome)
	router.Handle("/api/register", auth.BasicToken(http.HandlerFunc(account.Register)))
	router.Handle("/api/login", auth.BasicToken(http.HandlerFunc(account.Login)))
	router.HandleFunc("/api/refresh", account.RefreshToken)
	router.Handle("/api/profile", auth.UserMiddleware(http.HandlerFunc(account.UserProfile)))

	// Chat
	router.Handle("/api/chat/history", auth.UserMiddleware(http.HandlerFunc(chat.History)))
	router.Handle("/api/chat/history/messages", auth.UserMiddleware(http.HandlerFunc(chat.MessageHistory)))

	// Websocket
	router.Handle("/ws", auth.WebsocketAuthMiddleware(http.HandlerFunc(ws.WebSocketServer)))

	log.Println("Running on address: ", defaultServerAddress)
	if err := http.ListenAndServe(defaultServerAddress, router); err != http.ErrServerClosed {
		log.Println(err)
	}
}
