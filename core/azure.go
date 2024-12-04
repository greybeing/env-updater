package core

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Proceeding with system environment variables...")
	}
}

func UpdateAzureDevOpsFile(filename string, content []byte) error {
	// Retrieve Azure DevOps configuration
	pat := os.Getenv("AZURE_DEVOPS_PAT")
	org := os.Getenv("AZURE_DEVOPS_ORG")
	project := os.Getenv("AZURE_DEVOPS_PROJECT")

	// Validate environment variables
	if pat == "" {
		return fmt.Errorf("missing environment variable: AZURE_DEVOPS_PAT")
	}
	if org == "" {
		return fmt.Errorf("missing environment variable: AZURE_DEVOPS_ORG")
	}
	if project == "" {
		return fmt.Errorf("missing environment variable: AZURE_DEVOPS_PROJECT")
	}

	// Prepare Azure DevOps file upload URL
	apiURL := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1",
		org,
		project,
	)

	log.Printf("Uploading file to Azure DevOps: %s", filename)

	// Create HTTP request for file upload
	body := bytes.NewReader(content)
	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		return fmt.Errorf("failed to create upload request: %v", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Basic "+basicAuth("", pat))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	// Execute request
	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send upload request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("file upload failed with status code: %d", resp.StatusCode)
	}

	log.Printf("File successfully uploaded to Azure DevOps: %s", filename)
	return nil
}

// basicAuth is a helper for basic authentication encoding.
func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
