package ws

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/utils"
	"github.com/gorilla/websocket"
)

// WClient is a representation of a websocket client
type WClient struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func (c *WClient) putMsgInRoom() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *WClient) readMsgFromRoom(roomName string) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		pl := map[string][]byte{
			roomName: message,
		}
		c.hub.roomMessage <- pl
	}
}

// WebSocketServer is
func WebSocketServer(w http.ResponseWriter, req *http.Request) {

	// Retrieve user from context
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

	// Get for create room for two users to communicate
	roomName := getRoomName(int(user.ID), secondUserID)

	// Upgrade to websocket connection
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

	hub := newHub()
	go hub.run()

	client := &WClient{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	client.hub.roomName <- roomName

	go client.putMsgInRoom()
	go client.readMsgFromRoom(roomName)
}
