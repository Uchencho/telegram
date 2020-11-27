package ws

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Uchencho/telegram/db"
	"github.com/Uchencho/telegram/server/auth"
	"github.com/Uchencho/telegram/server/chat"
	"github.com/Uchencho/telegram/server/utils"
)

func (c *WClient) putMsgInRoom(user auth.User) {
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

			// Write to DB to store the chat
			msg := chat.Message{
				UserID:   int(user.ID),
				Username: user.FirstName,
				Thread:   c.Thread,
				Chatmsg:  string(message),
			}

			// Concurrently store the message
			go storeMessage(db.Db, msg)

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
		c.hub.roomName <- roomName
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

// WebSocketServer is a handler that connects a user for constant communication
func WebSocketServer(w http.ResponseWriter, req *http.Request) {

	// Retrieve user from context
	user := utils.GetUserFromRequestContext(w, req)

	urlValues := req.URL.Query()
	username := urlValues.Get("receiver_username")
	userID := urlValues.Get("receiver_id")
	if username == "" || userID == "" {
		utils.InvalidJSONResp(w, errors.New("Invalid query parameters passed"))
		return
	}

	secondUserID, err := strconv.Atoi(userID)
	if err != nil {
		utils.InvalidJSONResp(w, err)
		return
	}

	// Get for create room for two users to communicate
	roomName := getRoomName(int(user.ID), secondUserID)

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

	client := &WClient{hub: hub, conn: conn, Thread: threadID, send: make(chan []byte, 256)}
	client.hub.register <- client
	client.hub.roomName <- roomName

	go client.putMsgInRoom(user)
	go client.readMsgFromRoom(roomName)
}
