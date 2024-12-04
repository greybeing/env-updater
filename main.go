package main

import (
	"log"
	"os"
    "github.com/gin-gonic/gin"
	"env-updater/handlers"
)

func main() {
	// Configure logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Create Gin router
	router := gin.Default()

	// Register webhook endpoint
	router.POST("/webhook", handlers.HandleWebhook)

	// Determine server port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}