package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"env-updater/services"
)

type WebhookPayload struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Ref  string `json:"ref"`
	Commits []struct {
		Modified []string `json:"modified"`
	} `json:"commits"`
}

func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Failed to parse webhook payload", http.StatusBadRequest)
		return
	}

	var updatedFiles []string
	for _, commit := range payload.Commits {
		for _, file := range commit.Modified {
			if len(file) > 4 && file[:4] == "/" {
				updatedFiles = append(updatedFiles, file)
			}
		}
	}

	if len(updatedFiles) > 0 {
		go services.ProcessUpdatedFiles(updatedFiles)
	}

	fmt.Fprintln(w, "Webhook processed successfully.")
}
