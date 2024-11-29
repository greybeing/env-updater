package services

import (
	"fmt"
	"env-updater/core"
)

func ProcessUpdatedFiles(files []string) {
	repoOwner := "your-github-username"
	repoName := "your-repo-name"
	branch := "main"
	organization := "your-azure-org"
	project := "your-azure-project"

	for _, file := range files {
		fmt.Printf("Processing file: %s\n", file)

		// Step 1: Fetch the file
		filePath := core.FetchUpdatedFile(file, repoOwner, repoName, branch)

		// Step 2: Replace the file in Azure DevOps
		fileName := file // Use file name as it appears in the secure file library
		err := core.DeleteSecureFile(organization, project, fileName)
		if err != nil {
			fmt.Printf("Error deleting file: %v\n", err)
		}

		err = core.UploadSecureFile(organization, project, filePath, fileName)
		if err != nil {
			fmt.Printf("Error uploading file: %v\n", err)
		}
	}
}
