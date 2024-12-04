package services

import (
	"fmt"
	"log"
	"path/filepath"
    "env-updater/core"
	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Proceeding with system environment variables...")
	}
}


func ProcessWebhookEvent(webhookData map[string]interface{}) error {
	// Extract repository details
	repo, ok := webhookData["repository"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no repository data found")
	}

	fullName, ok := repo["full_name"].(string)
	if !ok {
		return fmt.Errorf("could not extract repository full name")
	}

	// Extract file details from push event
	files, ok := webhookData["commits"].([]interface{})
	if !ok || len(files) == 0 {
		return fmt.Errorf("no files found in webhook")
	}

	// Process each modified file
	for _, commitInterface := range files {
		commit, ok := commitInterface.(map[string]interface{})
		if !ok {
			continue
		}

		modifiedFiles, ok := commit["modified"].([]interface{})
		if !ok {
			continue
		}

		for _, fileInterface := range modifiedFiles {
			filename, ok := fileInterface.(string)
			if !ok {
				continue
			}

			// Fetch file content from GitHub
			fileContent, err := core.FetchFileFromGitHub(fullName, filename)
			if err != nil {
				log.Printf("GitHub file fetch error for %s: %v", filename, err)
				continue
			}

			// Update file in Azure DevOps
			if err := core.UpdateAzureDevOpsFile(filepath.Base(filename), fileContent); err != nil {
				log.Printf("Azure DevOps update error for %s: %v", filename, err)
				continue
			}

			log.Printf("Successfully processed file: %s", filename)
		}
	}

	return nil
}