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
	conn   *websocket.Conn
	send   chan []byte
	Thread int
	Room   chan map[string][]byte
}

func (c WClient) putMsgInRoom(sender, receiver, roomName string) {
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
			roomName: message,
			receiver: []byte(receiver),
		}
		c.Room <- payload
		log.Println("\n\nMessage put into room by:  ", sender, " Room name is: ", roomName, " message is: ", string(message))
	}
}

// readMsgFromRoom is
func (c WClient) readMsgFromRoom(sender, receiver, roomName string) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {

		select {
		case payload, ok := <-c.Room:

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

			if message, ok := payload[roomName]; ok {

				w, err := c.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}

				// Check which receiver should receive a message back
				if byteValue, ok := payload[receiver]; ok && string(byteValue) == receiver {
					log.Println("\n\nMessage read from room by:  ", receiver, " Room name is: ", roomName, "message is: ", string(message))

					// message = append(message, []byte(fmt.Sprintf(" Sent by %s and Should be received by %s ", sender, receiver))...)

					if err != nil {
						return
					}
					_, err = w.Write(message)
					if err != nil {
						log.Println(err)
					}

					log.Println("\n\nMessage sent by ", sender)
				}
			}
			continue

		case <-ticker.C:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Println(err)
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		default:
			continue
		}
		continue

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

	client := &WClient{
		conn: conn,
		send: make(chan []byte, 256),
		// Thread: threadID,
		Room: make(chan map[string][]byte, 5),
	}

	go client.putMsgInRoom(user.FirstName, username, roomName)
	go client.readMsgFromRoom(user.FirstName, username, roomName)
}
