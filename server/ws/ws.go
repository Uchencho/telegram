package ws

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte(`\n`)
	space   = []byte(` `)
	theMap  = map[string][]*WClient{}
)

const (
	maxMessageSize = 512
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	writeWait      = 10 * time.Second
)

// WClient is a representation of a websocket client
type WClient struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	Thread int
}

// Hub is a representation of a hub
type Hub struct {
	clients map[*WClient]bool

	// Register requests from the clients.
	register chan *WClient

	// Unregister requests from clients.
	unregister chan *WClient

	// rooms maps a room to a list of clients
	rooms map[string][]*WClient

	// roomMessage is a channel that sends a message into a specific room
	roomMessage chan map[string][]byte

	// roomName is the name of a room the client is connecting to
	roomName chan string
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
		roomMessage: make(chan map[string][]byte),
		register:    make(chan *WClient),
		unregister:  make(chan *WClient),
		clients:     make(map[*WClient]bool),
		rooms:       make(map[string][]*WClient),
		roomName:    make(chan string),
	}
}

func checkRoom(roomName string, client *WClient) map[string][]*WClient {
	clients, created := theMap[roomName]
	if created {
		for _, regClient := range clients {
			if regClient == client {
				return theMap
			}
		}
		clients = append(clients, client)
		theMap[roomName] = clients
		return theMap
	}
	theMap[roomName] = []*WClient{client}
	return theMap
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			roomName := <-h.roomName
			h.rooms = checkRoom(roomName, client)
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case incomingPL := <-h.roomMessage:
			// Go through each rooms in the hub, check which room a message was dropped in and send messages there
			for room, clients := range h.rooms {
				if message, ok := incomingPL[room]; ok {
					for _, client := range clients {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(h.clients, client)
						}
					}
				}
			}
		}
	}
}
