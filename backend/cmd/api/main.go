package main

import (
	"log"

	"github.com/novacart/graph-detective/internal/config"
	"github.com/novacart/graph-detective/internal/database"
	"github.com/novacart/graph-detective/internal/repository"
	"github.com/novacart/graph-detective/internal/router"
)

func main() {
	log.Println("Starting Graph Detective API server...")
	cfg := config.LoadConfig()

	// Connect to CognoDB
	db, err := database.Connect(cfg.CognoDBURI, cfg.CognoDBPassword)
	if err != nil {
		log.Printf("ERROR: Database connection failed (running in offline-ready mode): %v", err)
	} else {
		defer db.Close()
	}

	// Initialize Repository
	repo := repository.NewGraphRepository(db)

	// Setup separated router concerns
	r := router.SetupRouter(db, repo)

	log.Printf("Listening and serving HTTP on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to run HTTP server: %v", err)
	}
}
