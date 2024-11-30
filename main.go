package main

import (
	"log"
	"net/http"
	"os"
    "github.com/joho/godotenv"
	"env-updater/handlers"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	port := os.Getenv("PORT")
	// if port == "" {
	// 	port = "8080" // Default port.
	// }

	http.HandleFunc("/webhook", handlers.WebhookHandler)

	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
