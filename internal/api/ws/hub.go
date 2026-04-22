package ws

import (
	"log"
	"time"
)

// Message represents a payload sent over the WebSocket
type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// Hub maintains the set of active clients and broadcasts messages to the clients.
type Hub struct {
	// Registered clients, organized by room ID.
	rooms map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Message

	// Register requests from the clients.
	register chan *ClientSubscription

	// Unregister requests from clients.
	unregister chan *ClientSubscription
}

// ClientSubscription groups a client and the rooms it belongs to, 
// useful for registration and unregistration.
type ClientSubscription struct {
	client  *Client
	roomIDs []string
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Message),
		register:   make(chan *ClientSubscription),
		unregister: make(chan *ClientSubscription),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Run starts the Hub's main event loop.
// This is the core "single goroutine" handling all state changes to avoid mutexes.
func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.register:
			for _, roomID := range sub.roomIDs {
				connections := h.rooms[roomID]
				if connections == nil {
					connections = make(map[*Client]bool)
					h.rooms[roomID] = connections
				}
				connections[sub.client] = true
			}
			log.Printf("Client %s registered to %d rooms", sub.client.userID, len(sub.roomIDs))

		case sub := <-h.unregister:
			for _, roomID := range sub.roomIDs {
				connections := h.rooms[roomID]
				if connections != nil {
					if _, ok := connections[sub.client]; ok {
						delete(connections, sub.client)
						if len(connections) == 0 {
							delete(h.rooms, roomID)
						}
					}
				}
			}
			// Close the client's send channel to stop its writePump
			close(sub.client.send)
			log.Printf("Client %s unregistered", sub.client.userID)

		case message := <-h.broadcast:
			connections := h.rooms[message.RoomID]
			for client := range connections {
				select {
				case client.send <- message:
					// Message sent to client's channel
				default:
					// If the send channel is blocked (client is slow/dead), 
					// we close it and remove the client to prevent blocking the hub.
					close(client.send)
					delete(connections, client)
					if len(connections) == 0 {
						delete(h.rooms, message.RoomID)
					}
				}
			}
		}
	}
}
