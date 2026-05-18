package http

import (
	"database/sql"

	"github.com/Ahmad-Mosha/go-chat-api/internal/api/http/middleware"
	"github.com/Ahmad-Mosha/go-chat-api/internal/api/ws"
	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	"github.com/Ahmad-Mosha/go-chat-api/internal/repository/sqlite"
	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
)

// SetupRouter handles the dependency injection and route configuration
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	// --- Dependency Injection ---

	// Repositories
	userRepo := sqlite.NewUserRepository(db)
	roomRepo := sqlite.NewRoomRepository(db)
	msgRepo := sqlite.NewMessageRepository(db)

	// Services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg)
	chatService := service.NewChatService(roomRepo, msgRepo)

	// WebSockets
	wsHub := ws.NewHub()
	go wsHub.Run() // Start the Hub's event loop in the background

	// Handlers
	authHandler := NewAuthHandler(userService, authService)
	userHandler := NewUserHandler(userService)
	chatHandler := NewChatHandler(chatService)
	wsHandler := ws.NewHandler(wsHub, chatService)

	// --- Routes ---

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Auth Group (public)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", authHandler.SignUp)
		authGroup.POST("/login", authHandler.Login)
	}

	// Chat Group (protected by JWT middleware)
	api := r.Group("/api", middleware.AuthMiddleware(authService))
	{
		api.POST("/rooms", chatHandler.CreateRoom)
		api.GET("/rooms", chatHandler.GetUserRooms)
		api.GET("/rooms/:id", chatHandler.GetRoom)
		api.GET("/rooms/:id/members", chatHandler.GetRoomMembers)
		api.POST("/rooms/:id/messages", chatHandler.SendMessage)
		api.GET("/rooms/:id/messages", chatHandler.GetRoomMessages)

		api.GET("/users/lookup", userHandler.LookupUser)
		api.GET("/users/:id", userHandler.GetUser)

		api.GET("/ws", wsHandler.ServeWS)
	}

	return r
}

