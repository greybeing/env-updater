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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload WebhookPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	for _, commit := range payload.Commits {
		for _, file := range commit.Modified {
			if len(file) > 0 { // Process all files (in the root folder or elsewhere)
				log.Printf("Processing updated file: %s", file)
				err := services.ProcessUpdatedFile(file)
				if err != nil {
					log.Printf("Error processing file %s: %v", file, err)
				}
			}
		}
	}
    w.WriteHeader(http.StatusOK)

	fmt.Fprintln(w, "Webhook processed successfully")
}
