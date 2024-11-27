package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/webhook", handleWebhook)

	port := "8080"
	fmt.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var payload struct {
		Ref        string `json:"ref"`
		Repository struct {
			Name string `json:"name"`
		} `json:"repository"`
		Commits []struct {
			Added    []string `json:"added"`
			Modified []string `json:"modified"`
			Removed  []string `json:"removed"`
		} `json:"commits"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	fmt.Printf("Push event received for repository: %s on branch: %s\n", payload.Repository.Name, payload.Ref)

	// Check for updated files in the `env/` folder
	envFolder := "/"
	var updatedFiles []string

	for _, commit := range payload.Commits {
		for _, file := range append(commit.Added, commit.Modified...) {
			if len(file) >= len(envFolder) && file[:len(envFolder)] == envFolder {
				updatedFiles = append(updatedFiles, file)
			}
		}
	}

	if len(updatedFiles) > 0 {
		fmt.Printf("Updated files in '%s': %v\n", envFolder, updatedFiles)
		// Proceed to fetch and process the updated files
		go processUpdatedFiles(updatedFiles)
	} else {
		fmt.Println("No relevant files in the 'env/' folder were updated.")
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Webhook processed"))
}

func processUpdatedFiles(files []string) {
	for _, file := range files {
		fmt.Printf("Processing file: %s\n", file)
		// TODO: Implement logic to fetch and update the file
	}
}

