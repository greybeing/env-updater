package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

const azureDevOpsBaseURL = "https://dev.azure.com"

// DeleteSecureFile deletes a file from Azure DevOps secure files.
func DeleteSecureFile(organization, project, fileName string) error {
	pat := os.Getenv("AZURE_DEVOPS_PAT")
	if pat == "" {
		return fmt.Errorf("AZURE_DEVOPS_PAT is not set")
	}

	url := fmt.Sprintf("%s/%s/%s/_apis/distributedtask/securefiles/%s?api-version=7.1-preview.1", azureDevOpsBaseURL, organization, project, fileName)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}
	req.SetBasicAuth("", pat)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete file, status: %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("File deleted successfully.")
	return nil
}

// UploadSecureFile uploads a file to Azure DevOps secure files
func UploadSecureFile(organization, project, filePath, fileName string) error {
	pat := os.Getenv("AZURE_DEVOPS_PAT")
	if pat == "" {
		return fmt.Errorf("AZURE_DEVOPS_PAT is not set")
	}

	url := fmt.Sprintf("%s/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1", azureDevOpsBaseURL, organization, project)

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(fileContent))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.SetBasicAuth("", pat)
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload file, status: %d, response: %s", resp.StatusCode, string(body))
	}

	fmt.Println("File uploaded successfully.")
	return nil
}
