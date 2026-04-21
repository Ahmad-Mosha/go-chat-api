package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Ahmad-Mosha/go-chat-api/internal/api/http" // Import aliased via internal package name
	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Database Connection Pool
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	dbPool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer dbPool.Close()

	// 3. Setup Router (This is where all the wiring now happens)
	router := http.SetupRouter(dbPool, cfg)

	// 4. Start Server
	fmt.Printf("Server starting on port %s...\n", cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
