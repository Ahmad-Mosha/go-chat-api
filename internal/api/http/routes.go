package http

import (
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
	
	// Services
	userService := service.NewUserService(userRepo)
	authService := service.NewAuthService(userRepo, cfg)
	
	// Handlers
	authHandler := NewAuthHandler(userService, authService)

	// --- Routes ---

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Auth Group
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", authHandler.SignUp)
		authGroup.POST("/login", authHandler.Login)
	}

	return r
}
