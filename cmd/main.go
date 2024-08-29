package main

import (
	"log"

	"github.com/yeboahd24/subscription-stripe/config"
	"github.com/yeboahd24/subscription-stripe/database"
	"github.com/yeboahd24/subscription-stripe/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set up Gin router
	r := gin.Default()

	// Set up routes
	routes.SetupRoutes(r, db)

	// Start the server
	log.Printf("Server starting on %s", cfg.ServerAddress)
	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
