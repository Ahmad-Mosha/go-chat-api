package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles HTTP requests for rooms and messages.
type ChatHandler struct {
	chatService *service.ChatService
}

// NewChatHandler creates a new instance of the chat handler.
func NewChatHandler(cs *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: cs}
}

// --- Request DTOs ---

type createRoomRequest struct {
	Name      string   `json:"name"`
	IsGroup   bool     `json:"is_group"`
	MemberIDs []string `json:"member_ids" binding:"required,min=1"`
}

type sendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// --- Handlers ---

// CreateRoom handles POST /rooms
func (h *ChatHandler) CreateRoom(c *gin.Context) {
	userID := c.GetString("user_id")

	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.chatService.CreateRoom(userID, req.Name, req.IsGroup, req.MemberIDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrCannotChatSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"room": room,
	})
}

// GetUserRooms handles GET /rooms
func (h *ChatHandler) GetUserRooms(c *gin.Context) {
	userID := c.GetString("user_id")

	rooms, err := h.chatService.GetUserRooms(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}

// GetRoom handles GET /rooms/:id
func (h *ChatHandler) GetRoom(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	room, err := h.chatService.GetRoom(roomID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotRoomMember):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrRoomNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room": room,
	})
}

// SendMessage handles POST /rooms/:id/messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	msg, err := h.chatService.SendMessage(roomID, userID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrNotRoomMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": msg,
	})
}

// GetRoomMessages handles GET /rooms/:id/messages
func (h *ChatHandler) GetRoomMessages(c *gin.Context) {
	userID := c.GetString("user_id")
	roomID := c.Param("id")

	// Parse optional query params: ?limit=50&offset=0
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.chatService.GetRoomMessages(roomID, userID, limit, offset)
	if err != nil {
		if errors.Is(err, service.ErrNotRoomMember) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}
