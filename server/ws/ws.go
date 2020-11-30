package ws

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline        = []byte(`\n`)
	space          = []byte(` `)
	roomAndClients = map[string][]*WClient{}
	globalHUB      = newHub()
	lock           sync.Mutex
)

const (
	maxMessageSize = 512
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
)

// WClient is a representation of a websocket client
type WClient struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan wsPayload
	Thread   int
	userName string
	userID   int
	roomName string
}

type wsPayload struct {
	sender   *WClient
	roomName string
	message  []byte
}

// Hub is a representation of a hub
type Hub struct {

	// Register requests from the clients.
	register chan *WClient

	// Unregister requests from clients.
	unregister chan *WClient

	// roomMessage is a channel that sends a wsPayload into a specific room
	roomMessage chan wsPayload
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

func newHub() *Hub {

	return &Hub{
		roomMessage: make(chan wsPayload),
		register:    make(chan *WClient),
		unregister:  make(chan *WClient),
	}
}

// checkRoom checks if a room has been created and if the client has been put in the room
func checkRoom(roomName string, client *WClient) {
	clients, created := roomAndClients[roomName]
	if created {
		for _, regClient := range clients {
			if regClient == client {
				return
			}
		}
		clients = append(clients, client)
		roomAndClients[roomName] = clients
		return
	}
	roomAndClients[roomName] = []*WClient{client}
	return
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:

			lock.Lock()
			checkRoom(client.roomName, client)
			lock.Unlock()

		case client := <-h.unregister:
			cleanRoomAndClients(client.roomName, client)

		case incomingPL := <-h.roomMessage:

			lock.Lock()
			if clients, ok := roomAndClients[incomingPL.roomName]; ok {
				for _, client := range clients {
					if incomingPL.sender != client {
						select {
						case client.send <- incomingPL:
						default:
							cleanRoomAndClients(incomingPL.roomName, client)
						}
					}
				}
			}
			lock.Unlock()
		}
	}
}

func cleanRoomAndClients(roomName string, c *WClient) {
	clients, found := roomAndClients[roomName]
	if !found {
		log.Println("Room does not exist... This should never happen by the way")
		return
	}

	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			close(client.send)
		}
	}

	if len(clients) == 0 {
		delete(roomAndClients, roomName)
	}
}
