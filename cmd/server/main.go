package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Ahmad-Mosha/go-chat-api/internal/api/http" // Import aliased via internal package name
	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	_ "github.com/mattn/go-sqlite3" // Required for local SQLite file fallback by libsql
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Database Connection (Turso / libSQL)
	// If connecting to a remote Turso DB, the auth token is needed as a query param.
	connString := cfg.TursoURL
	if cfg.TursoAuthToken != "" {
		connString = fmt.Sprintf("%s?authToken=%s", cfg.TursoURL, cfg.TursoAuthToken)
	}

	db, err := sql.Open("libsql", connString)
	if err != nil {
		log.Fatalf("Unable to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	// 3. Setup Router (This is where all the wiring now happens)
	router := http.SetupRouter(db, cfg)

	// 4. Start Server
	fmt.Printf("Server starting in %s environment on port %s...\n", cfg.AppEnv, cfg.ServerPort)
	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
