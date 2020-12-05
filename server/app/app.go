package app

import (
	"net/http"

	"github.com/Uchencho/telegram/server/account"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/utils"
	"github.com/Uchencho/telegram/server/ws"
	"github.com/gorilla/mux"
)

const defaultServerAddress = "127.0.0.1:8000"

// App is a representation of the set of functionalities for the app
type App struct {
	RegisterHandler       http.HandlerFunc
	LoginHandler          http.HandlerFunc
	RefreshTokenHandler   http.HandlerFunc
	UserProfileHandler    http.HandlerFunc
	ChatHistoryHandler    http.HandlerFunc
	MessageHistoryHandler http.HandlerFunc
	WebsocketHandler      http.HandlerFunc
}

// NewApp returns a new application
func NewApp() App {
	regHandler := auth.BasicToken(http.HandlerFunc(account.Register))
	loginHandler := auth.BasicToken(http.HandlerFunc(account.Login))
	refreshT := account.RefreshToken
	userProfileHandler := auth.UserMiddleware(http.HandlerFunc(account.UserProfile))
	chatHistoryHandler := auth.UserMiddleware(http.HandlerFunc(chat.History))
	messageHistoryHandler := auth.UserMiddleware(http.HandlerFunc(chat.MessageHistory))
	wsHandler := auth.WebsocketAuthMiddleware(http.HandlerFunc(ws.WebSocketServer))

	return App{
		RegisterHandler:       regHandler,
		LoginHandler:          loginHandler,
		RefreshTokenHandler:   refreshT,
		UserProfileHandler:    userProfileHandler,
		ChatHistoryHandler:    chatHistoryHandler,
		MessageHistoryHandler: messageHistoryHandler,
		WebsocketHandler:      wsHandler,
	}
}

// Handler returns the main Handler for the application
func (a *App) Handler() http.HandlerFunc {
	router := mux.NewRouter()
	router.NotFoundHandler = auth.BasicToken(http.HandlerFunc(utils.NotAvailabe))

	// router.HandleFunc("/", serveHome)
	router.Handle("/api/register", a.RegisterHandler)
	router.Handle("/api/login", a.LoginHandler)
	router.HandleFunc("/api/refresh", a.RefreshTokenHandler)
	router.Handle("/api/profile", a.UserProfileHandler)

	// Chat
	router.Handle("/api/chat/history", a.ChatHistoryHandler)
	router.Handle("/api/chat/history/messages", a.MessageHistoryHandler)

	// Websocket
	router.Handle("/ws", a.WebsocketHandler)
	h := http.HandlerFunc(router.ServeHTTP)
	return h
}
