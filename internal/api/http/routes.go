package http

import (
	"github.com/Ahmad-Mosha/go-chat-api/internal/api/http/middleware"
	"github.com/Ahmad-Mosha/go-chat-api/internal/api/ws"
	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	"github.com/Ahmad-Mosha/go-chat-api/internal/repository/postgres"
	"github.com/Ahmad-Mosha/go-chat-api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter handles the dependency injection and route configuration
func SetupRouter(dbPool *pgxpool.Pool, cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// --- Dependency Injection ---

	// Repositories
	userRepo := postgres.NewUserRepository(dbPool)
	roomRepo := postgres.NewRoomRepository(dbPool)
	msgRepo := postgres.NewMessageRepository(dbPool)

	// Services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg)
	chatService := service.NewChatService(roomRepo, msgRepo)

	// WebSockets
	wsHub := ws.NewHub()
	go wsHub.Run() // Start the Hub's event loop in the background

	// Handlers
	authHandler := NewAuthHandler(userService, authService)
	// Optionally, we could pass the Hub to chatHandler if it needs to broadcast from REST endpoints.
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
		api.POST("/rooms/:id/messages", chatHandler.SendMessage)
		api.GET("/rooms/:id/messages", chatHandler.GetRoomMessages)
		
		// WebSocket endpoint (also protected by JWT middleware via query or header depending on client)
		// For a real app, clients often pass JWT in a query param since browser JS WS API can't set headers easily.
		// For now, we will use the same middleware which expects the "Authorization" header.
		api.GET("/ws", wsHandler.ServeWS)
	}

	return r
}

