package ws

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/utils"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte(`\n`)
	space   = []byte(` `)
)

const (
	maxMessageSize = 512
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
)

// Client is a representation of a websocket client
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	Thread int
	Room   string
}

// Hub is a representation of a hub
type Hub struct {
	Clients    map[*Client]bool
	Payload    chan map[string][]byte
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

func getRoomName(userOneID, userTwoID int) (roomName string) {

	userOneString := strconv.Itoa(userOneID)
	userTwoString := strconv.Itoa(userTwoID)
	if userOneID <= userTwoID {
		roomName = userOneString + "_" + userTwoString
	} else {
		roomName = userTwoString + "_" + userOneString
	}
	return fmt.Sprintf("room_%s", roomName)
}

// NewHub Creates a new hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Payload:    make(chan map[string][]byte),
	}
}

// Run Checks the status of the hub and sends the appropraite signal to the channel
func (h *Hub) Run() {
	for {
		log.Println("Details of the hub is: ", h, "Payload is: ", h.Payload, "Register: ", h.Register, h.Unregister, "Clients: ", h.Clients)
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case p := <-h.Payload:
			log.Println("\n\nHub things here: ", time.Now(), ">>>>>>", p)
			for client := range h.Clients {
				message, present := p[client.Room]
				if present {
					log.Println("\n\nFound a room ", time.Now(), ">>>>", client.Room)
					client.send <- message
				} else {
					close(client.send)
					delete(h.Clients, client)
				}
			}
		case message := <-h.Broadcast:
			// Retrieve the id that the message is going to
			for client := range h.Clients {
				select {
				case client.send <- message:
					log.Println("\n\nBroadcast things here: ", time.Now())
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// Send messages to the hub
func (c *Client) sendMessage(user auth.User) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			log.Println("\n\nI just got called", time.Now())
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println(err)
			}
			if !ok {
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Println(err)
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, err = w.Write(message)
			if err != nil {
				log.Println(err)
			}

			// Write to DB to store the chat
			// msg := chat.Message{
			// 	UserID:   int(user.ID),
			// 	Username: user.FirstName,
			// 	Thread:   c.Thread,
			// 	Chatmsg:  string(message),
			// }
			// err = storeMessage(db.Db, msg)
			// if err != nil {
			// 	log.Println("\n\n", err)
			// }

			n := len(c.send)
			for i := 0; i < n; i++ {
				_, err = w.Write(newline)
				if err != nil {
					log.Println(err)
				}

				_, err = w.Write(<-c.send)
				if err != nil {
					log.Println(err)
				}
			}

			if err := w.Close(); err != nil {
				log.Println(err)
				return
			}

		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println(err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Read any messages sent
func (c *Client) readMessage() {

	defer func() {
		c.hub.Unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Println(err)
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			log.Println(err)
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("error: ", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		// Put in message into a room
		payload := map[string][]byte{
			c.Room: message,
		}
		// c.hub.Broadcast <- message
		log.Println("\n\nMessage put into hub:  ", time.Now(), "Room name is: ", c.Room, "message is: ", string(message))
		c.hub.Payload <- payload
	}
}

// ChatServer Takes in the http request, upgrade it to a websocket request and spring up two go routines
func ChatServer(w http.ResponseWriter, req *http.Request) {

	const userKey auth.Key = "user"
	user, ok := req.Context().Value(userKey).(auth.User)
	if !ok {
		utils.InternalIssues(w, errors.New("Cannot decode context from middleware"))
		return
	}

	urlValues := req.URL.Query()
	username := urlValues.Get("receiver_username")
	userID := urlValues.Get("receiver_id")
	if username == "" || userID == "" {
		utils.InvalidJsonResp(w, errors.New("Invalid query parameters passed"))
		return
	}

	secondUserID, err := strconv.Atoi(userID)
	if err != nil {
		utils.InvalidJsonResp(w, err)
		return
	}

	threadInput := chat.Thread{
		FirstUserID:    int(user.ID),
		FirstUsername:  user.FirstName,
		SecondUserID:   secondUserID,
		SecondUsername: username,
	}

	threadID, err := getOrCreateThread(db.Db, threadInput)
	if err != nil {
		utils.InternalIssues(w, err)
		return
	}

	hub := NewHub()
	go hub.Run()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		Thread: threadID,
		Room:   getRoomName(int(user.ID), secondUserID),
	}
	client.hub.Register <- client

	go client.readMessage()
	go client.sendMessage(user)
}

/*
Connect to a specific room
Send message
Read message from a specific room

Case when the message is being put in a room
Find the client for that room and send him a message
*/
