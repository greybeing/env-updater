package core

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func init() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found. Proceeding with system environment variables...")
    }
}

// UpdateAzureDevOpsFile deletes the existing file and uploads a new version, using dynamically selected project
func UpdateAzureDevOpsFile(ctx context.Context, filename string) error {
    // Retrieve Azure DevOps configuration 
    pat := os.Getenv("AZURE_DEVOPS_PAT")
    org := os.Getenv("AZURE_DEVOPS_ORG")
    project := os.Getenv("AZURE_DEVOPS_PROJECT")

    log.Printf("AZURE_DEVOPS_PAT: [REDACTED]")
    log.Printf("AZURE_DEVOPS_ORG: %s", org)
    log.Printf("AZURE_DEVOPS_PROJECT: %s", project)

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

    // Define the path to "security.txt"
    securityFilePath := "security.txt"

    // Ensure the security.txt file exists
    if _, err := os.Stat(securityFilePath); os.IsNotExist(err) {
        return fmt.Errorf("security.txt file does not exist in the root directory")
    }
    log.Println("Confirmed security.txt exists")

    // Read the content of "security.txt"
    content, err := os.ReadFile(securityFilePath)
    if err != nil {
        return fmt.Errorf("failed to read security.txt: %v", err)
    }
    log.Printf("Successfully read security.txt with %d bytes", len(content))

    // Check if the file exists before attempting to delete
    fileExists, secureFileId, err := checkFileExists(ctx, filename, pat, org, project)
    if err != nil {
        return fmt.Errorf("error checking if file exists: %v", err)
    }

    if fileExists {
        // Delete the existing file
        deleteErr := deleteFile(ctx, secureFileId, pat, org, project)
        if deleteErr != nil {
            return fmt.Errorf("error deleting file: %v", deleteErr)
        }
        log.Printf("File %s was successfully deleted", filename)
    } else {
        log.Printf("File %s not found, proceeding to upload", filename)
    }

    // Now, upload the new version
    apiURL := fmt.Sprintf(
        "https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1&name=%s",
        org,
        project,
        filename,
    )

    log.Printf("Preparing to upload file %s to Azure DevOps in project %s", filename, project)
    log.Printf("Request URL: %s", apiURL)

    // Create HTTP request for file upload with context
    req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(content))
    if err != nil {
        return fmt.Errorf("failed to create upload request: %v", err)
    }

    // Set headers
    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))
    req.Header.Set("Content-Type", "application/octet-stream")
    log.Printf("Request Headers: %+v", req.Header)

    // Execute request
    client := &http.Client{
        Timeout: 30 * time.Second, // Set a timeout
    }
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send upload request: %v", err)
    }
    defer resp.Body.Close()

    // Read full response body for detailed logging
    bodyBytes, _ := io.ReadAll(resp.Body)

    // Check response status
    if resp.StatusCode != http.StatusCreated {
        log.Printf("Azure DevOps API Response - Status: %d, Body: %s", resp.StatusCode, string(bodyBytes))
        return fmt.Errorf("file upload failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
    }

    log.Printf("File %s successfully uploaded to Azure DevOps with status: %d", filename, resp.StatusCode)

    return nil
}

// checkFileExists checks if a file with given name exists in Azure DevOps Secure Files
func checkFileExists(ctx context.Context, filename, pat, org, project string) (bool, string, error) {
    apiURL := fmt.Sprintf(
        "https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.1-preview.1",
        org,
        project,
    )

    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if err != nil {
        return false, "", fmt.Errorf("failed to create get request for file check: %v", err)
    }
    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))

    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return false, "", fmt.Errorf("failed to send get request for file check: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return false, "", fmt.Errorf("failed to get secure files list: status code %d", resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return false, "", fmt.Errorf("failed to read response body: %v", err)
    }

    var secureFiles struct {
        Value []struct {
            Id   string `json:"id"`
            Name string `json:"name"`
        } `json:"value"`
    }

    if err := json.Unmarshal(body, &secureFiles); err != nil {
        return false, "", fmt.Errorf("failed to unmarshal JSON response: %v", err)
    }

    for _, file := range secureFiles.Value {
        if file.Name == filename {
            return true, file.Id, nil
        }
    }

    return false, "", nil
}

// deleteFile deletes a file from Azure DevOps Secure Files library
func deleteFile(ctx context.Context, secureFileId, pat, org, project string) error {
    deleteURL := fmt.Sprintf(
        "https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles/%s?api-version=7.1-preview.1",
        org,
        project,
        secureFileId,
    )

    req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
    if err != nil {
        return fmt.Errorf("failed to create delete request: %v", err)
    }
    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))

    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send delete request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNoContent {
        return fmt.Errorf("failed to delete file: status code %d", resp.StatusCode)
    }

    return nil
}