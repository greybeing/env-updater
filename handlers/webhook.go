package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
    "env-updater/services"
)

type WebhookPayload struct {
	Commits []struct {
		Modified []string `json:"modified"`
	} `json:"commits"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Webhook triggered")

	// Ensure only POST requests are allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Ensure Content-Type is application/json
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Unsupported content type", http.StatusUnsupportedMediaType)
		return
	}

	// Read and log the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read request body: %v", err)
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	log.Printf("Received payload: %s", body)

	// Parse the JSON payload
	var payload WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		log.Printf("Invalid JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Process the modified files
	for _, commit := range payload.Commits {
		for _, file := range commit.Modified {
			if len(file) > 0 { // Process all files
				log.Printf("Processing updated file: %s", file)
				err := services.ProcessUpdatedFile(file)
				if err != nil {
					log.Printf("Error processing file %s: %v", file, err)
				}
			}
		}
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Webhook processed successfully")
	log.Printf("Webhook processed successfully")
}
