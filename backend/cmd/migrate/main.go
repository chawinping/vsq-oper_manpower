package main

import (
	"log"
	"os"

	"vsq-oper-manpower/backend/internal/config"
	"vsq-oper-manpower/backend/internal/repositories/postgres"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Running migrations...")

	// Run migrations
	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
		os.Exit(1)
	}

	log.Println("Migrations completed successfully!")
}
