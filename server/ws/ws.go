package ws

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Uchencho/telegram/server/auth"
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
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub is a representation of a hub
type Hub struct {
	Clients    map[*Client]bool
	Room       string
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// NewHub Creates a new hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

// Run Checks the status of the hub and sends the appropraite signal to the channel
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			// Retrieve the id that the message is going to
			for client := range h.Clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

// Send messages to the hub
func (c *Client) sendMessage() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
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
		c.hub.Broadcast <- message
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
	log.Println("Retrieving user deatils, ", user)

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
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.Register <- client

	go client.readMessage()
	go client.sendMessage()
}

/*
	Retrieve the user id, the logged in user is trying to send a message to
	Pass in the id to the function informing it which room to send the message to

*/
