package ws

import (
	"log"
	"net/http"

	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// In a real application, CheckOrigin should validate the origin
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler handles websocket related HTTP requests
type Handler struct {
	hub         *Hub
	chatService *service.ChatService
}

func NewHandler(hub *Hub, chatService *service.ChatService) *Handler {
	return &Handler{
		hub:         hub,
		chatService: chatService,
	}
}

// ServeWS upgrades the HTTP connection to a WebSocket and registers the client.
func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Get the user from the context (set by AuthMiddleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// 2. Upgrade the HTTP connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}

	// 3. Fetch the user's rooms so we can subscribe them to the correct hub channels.
	rooms, err := h.chatService.GetUserRooms(userID)
	if err != nil {
		log.Printf("failed to get user rooms: %v", err)
		conn.Close()
		return
	}

	var roomIDs []string
	for _, room := range rooms {
		roomIDs = append(roomIDs, room.ID)
	}

	// 4. Create the Client instance
	client := &Client{
		hub:     h.hub,
		userID:  userID,
		conn:    conn,
		send:    make(chan *Message, 256),
		roomIDs: roomIDs,
	}

	// 5. Register with the hub
	h.hub.register <- &ClientSubscription{
		client:  client,
		roomIDs: roomIDs,
	}

	// 6. Start the read and write pumps in separate goroutines
	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
