package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// SecureFile represents a secure file in Azure DevOps.
type SecureFile struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LoadEnv loads environment variables from a .env file.
func LoadEnv() error {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		return fmt.Errorf("failed to load .env file: %w", err)
	}
	return nil
}

// GetSecureFiles retrieves all secure files in an Azure DevOps project.
func GetSecureFiles() ([]SecureFile, error) {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		return nil, err
	}

	organization := os.Getenv("AZURE_ORGANIZATION")
	project := os.Getenv("AZURE_PROJECT")
	pat := os.Getenv("AZURE_DEVOPS_PAT")

	apiURL := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1",
		organization,
		project,
	)

	log.Printf("Fetching secure files from Azure DevOps: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth("", pat))
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending request to fetch secure files: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to fetch secure files. Status: %d, Response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("failed to fetch secure files, status: %d", resp.StatusCode)
	}

	var result struct {
		Value []SecureFile `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding secure files response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	log.Printf("Fetched %d secure files successfully.", len(result.Value))
	return result.Value, nil
}

// FindSecureFileByName finds a secure file by its name.
func FindSecureFileByName(files []SecureFile, fileName string) (*SecureFile, error) {
	log.Printf("Searching for secure file with name: %s", fileName)
	for _, file := range files {
		if strings.EqualFold(file.Name, fileName) { // Case-insensitive comparison
			log.Printf("Secure file found: %s (ID: %s)", file.Name, file.ID)
			return &file, nil
		}
	}
	log.Printf("Secure file not found: %s", fileName)
	return nil, fmt.Errorf("secure file with name %s not found", fileName)
}

// DeleteSecureFile deletes a secure file by its ID.
func DeleteSecureFile(fileID string) error {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		return err
	}

	organization := os.Getenv("AZURE_ORGANIZATION")
	project := os.Getenv("AZURE_PROJECT")
	pat := os.Getenv("AZURE_DEVOPS_PAT")

	apiURL := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles/%s?api-version=7.1-preview.1",
		organization,
		project,
		fileID,
	)

	log.Printf("Deleting secure file with ID: %s", fileID)

	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		log.Printf("Failed to create delete request: %v", err)
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+basicAuth("", pat))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending delete request: %v", err)
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to delete secure file. Status: %d, Response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("failed to delete secure file, status: %d", resp.StatusCode)
	}

	log.Printf("Secure file deleted successfully: %s", fileID)
	return nil
}

// UploadSecureFile uploads a file to Azure DevOps.
func UploadSecureFile(localFilePath, fileName string) error {
	// Load environment variables
	if err := LoadEnv(); err != nil {
		return err
	}

	organization := os.Getenv("AZURE_ORGANIZATION")
	project := os.Getenv("AZURE_PROJECT")
	pat := os.Getenv("AZURE_DEVOPS_PAT")

	apiURL := fmt.Sprintf(
		"https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1",
		organization,
		project,
	)

	log.Printf("Uploading file to Azure DevOps: %s (Local Path: %s)", fileName, localFilePath)

	file, err := os.Open(localFilePath)
	if err != nil {
		log.Printf("Error opening file for upload: %v", err)
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(localFilePath))
	if err != nil {
		log.Printf("Error creating form file: %v", err)
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		log.Printf("Error copying file content to form: %v", err)
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", apiURL, body)
	if err != nil {
		log.Printf("Error creating upload request: %v", err)
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+basicAuth("", pat))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending upload request: %v", err)
		return fmt.Errorf("failed to send upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Failed to upload file. Status: %d, Response: %s", resp.StatusCode, string(body))
		return fmt.Errorf("failed to upload file, status: %d", resp.StatusCode)
	}

	log.Printf("File uploaded successfully to Azure DevOps: %s", fileName)
	return nil
}

// basicAuth is a helper for basic authentication encoding.
func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}
