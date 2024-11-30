package services

import (
	"fmt"
	"os"
    "env-updater/core"
	"log"
)

func ProcessUpdatedFile(filePath string) error {
	repo := os.Getenv("GITHUB_REPO")
	branch := "main" // You can make this dynamic if needed

	// Fetch the updated file from GitHub
	fileContent, err := core.FetchUpdatedFile(repo, filePath, branch)
	if err != nil {
		return fmt.Errorf("failed to fetch updated file: %w", err)
	}

	// Write the file content to a temporary local file
	tempFilePath := "/tmp/" + filePath
	err = os.WriteFile(tempFilePath, fileContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file to temp path: %w", err)
	}

	// Upload the file to Azure DevOps
	organization := os.Getenv("AZURE_ORGANIZATION")
	project := os.Getenv("AZURE_PROJECT")

	err = core.DeleteSecureFile(organization, project, filePath)
	if err != nil {
		return fmt.Errorf("failed to delete existing secure file: %w", err)
	}

	err = core.UploadSecureFile(organization, project, tempFilePath, filePath)
	if err != nil {
		return fmt.Errorf("failed to upload secure file: %w", err)
	}

	fmt.Println("File processed and updated in Azure DevOps successfully.")
	log.Printf("File uploaded successfully.")
	return nil
}
