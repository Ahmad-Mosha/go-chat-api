package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/Ahmad-Mosha/go-chat-api/internal/config"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize Database Connection
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

	fmt.Println("Connected to Turso database successfully. Running migrations...")

	// 3. Find and read migration files
	files, err := filepath.Glob("migrations/*.up.sql")
	if err != nil {
		log.Fatalf("Failed to find migration files: %v", err)
	}

	// Ensure files are executed in order (000001, 000002, etc.)
	sort.Strings(files)

	// 4. Execute each migration file
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file, err)
		}

		fmt.Printf("Applying %s...\n", filepath.Base(file))
		_, err = db.Exec(string(content))
		if err != nil {
			log.Fatalf("Failed to apply migration %s: %v", file, err)
		}
	}

	fmt.Println("All migrations applied successfully!")
}
