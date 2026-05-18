package http

import (
	"errors"
	"net/http"

	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(us *service.UserService) *UserHandler {
	return &UserHandler{userService: us}
}

type userPublicResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

func toPublicUser(user *service.UserProfile) userPublicResponse {
	return userPublicResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}
}

// GetUser handles GET /api/users/:id
func (h *UserHandler) GetUser(c *gin.Context) {
	profile, err := h.userService.GetProfile(c.Param("id"))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toPublicUser(profile)})
}

// LookupUser handles GET /api/users/lookup?email=
func (h *UserHandler) LookupUser(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
		return
	}

	profile, err := h.userService.GetByEmail(email)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toPublicUser(profile)})
}
