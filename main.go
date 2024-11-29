package main

import (
	"log"
	"net/http"
    "env-updater/handlers"
)

func main() {
	http.HandleFunc("/webhook", handlers.WebhookHandler)

	log.Println("Starting server on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
