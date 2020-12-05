package app

import (
	"database/sql"
	"net/http"

	"github.com/Uchencho/telegram/server/account"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/database"
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

// Option is a representation of configurable options for the app
type Option struct {
	GetUserLogin      database.RetrieveUserLoginDetailsFunc
	UpdateUserDetails database.UpdateUserDetailsFunc
	AddNewUser        database.AddUserToDBFunc
	GetUserThreads    database.RetrieveUserThreadsFunc
	GetMessages       database.RetrieveMessagesFunc
	InsertMsg         database.InsertMessageFunc
	GetThread         database.GetorCreateThreadFunc
}

// NewApp returns a new application
func NewApp(provideDB *sql.DB) App {

	o := Option{
		GetUserLogin:      database.GetUserLogin(provideDB),
		UpdateUserDetails: database.UpdateUserRecord(provideDB),
		AddNewUser:        database.AddRecordToUserTable(provideDB),
		GetUserThreads:    database.ChatThreadsByUser(provideDB),
		GetMessages:       database.GetMessages(provideDB),
		InsertMsg:         database.StoreMessage(provideDB),
		GetThread:         database.GetOrCreateThread(provideDB),
	}

	regHandler := auth.BasicToken(http.HandlerFunc(account.Register(o.AddNewUser)))
	loginHandler := auth.BasicToken(http.HandlerFunc(account.Login(o.GetUserLogin)))
	refreshT := account.RefreshToken()
	userProfileHandler := auth.UserMiddleware(o.GetUserLogin, http.HandlerFunc(account.UserProfile(o.UpdateUserDetails)))
	chatHistoryHandler := auth.UserMiddleware(o.GetUserLogin, http.HandlerFunc(chat.History(o.GetUserThreads)))
	messageHistoryHandler := auth.UserMiddleware(o.GetUserLogin, http.HandlerFunc(chat.MessageHistory(o.GetMessages)))
	wsHandler := auth.WebsocketAuthMiddleware(o.GetUserLogin, http.HandlerFunc(ws.WebSocketServer(o.InsertMsg, o.GetThread)))

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
