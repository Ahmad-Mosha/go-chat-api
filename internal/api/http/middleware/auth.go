package middleware

import (
	"net/http"
	"strings"

	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT token and sets the user_id in the context.
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// 2. Extract the token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be: Bearer <token>"})
			c.Abort()
			return
		}

		// 3. Validate the token
		userID, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// 4. Set user_id in context for downstream handlers
		c.Set("user_id", userID)
		c.Next()
	}
}
