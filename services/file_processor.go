package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"env-updater/core"
)

// ProcessUpdatedFile processes and updates the secure file in Azure DevOps.
func ProcessUpdatedFile(filePath string) error {
	// Log the start of the process
	log.Printf("Starting processing for file: %s", filePath)

	// Load environment variables
	if err := core.LoadEnv(); err != nil {
		log.Printf("Error loading environment variables: %v", err)
		return fmt.Errorf("failed to load environment variables: %w", err)
	}

	// Fetch required environment variables
	repo := os.Getenv("GITHUB_REPO")
	branch := "main" // This can be made dynamic if needed

	// Validate required environment variables
	if repo == "" {
		log.Fatal("GITHUB_REPO environment variable must be set")
	}

	// Fetch the updated file from GitHub
	log.Printf("Fetching updated file from GitHub repository: %s, branch: %s", repo, branch)
	fileContent, err := core.FetchUpdatedFile(repo, filePath, branch)
	if err != nil {
		log.Printf("Error fetching updated file: %v", err)
		return fmt.Errorf("failed to fetch updated file: %w", err)
	}

	// Write the file content to a temporary local file
	tempFilePath := filepath.Join(os.TempDir(), filepath.Base(filePath))
	log.Printf("Writing file to temporary path: %s", tempFilePath)
	err = os.WriteFile(tempFilePath, fileContent, 0644)
	if err != nil {
		log.Printf("Error writing file to temporary path: %v", err)
		return fmt.Errorf("failed to write file to temp path: %w", err)
	}

	// Get the list of secure files in Azure DevOps
	log.Println("Fetching secure files from Azure DevOps")
	secureFiles, err := core.GetSecureFiles()
	if err != nil {
		log.Printf("Error fetching secure files: %v", err)
		return fmt.Errorf("failed to fetch secure files: %w", err)
	}

	// Find the secure file by its name
	log.Printf("Looking for secure file with name: %s", filePath)
	file, err := core.FindSecureFileByName(secureFiles, filepath.Base(filePath))
	if err != nil {
		log.Printf("Error finding secure file: %v", err)
		return fmt.Errorf("failed to find secure file: %w", err)
	}

	// Delete the secure file
	log.Printf("Deleting secure file with ID: %s", file.ID)
	err = core.DeleteSecureFile(file.ID)
	if err != nil {
		log.Printf("Error deleting secure file: %v", err)
		return fmt.Errorf("failed to delete existing secure file: %w", err)
	}

	// Upload the updated file
	log.Printf("Uploading updated file to Azure DevOps: %s", filePath)
	err = core.UploadSecureFile(tempFilePath, filepath.Base(filePath))
	if err != nil {
		log.Printf("Error uploading secure file: %v", err)
		return fmt.Errorf("failed to upload secure file: %w", err)
	}

	log.Printf("File processed and updated in Azure DevOps successfully: %s", filePath)
	return nil
}
