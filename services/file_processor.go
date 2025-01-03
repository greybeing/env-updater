package services

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
    "path/filepath"
    "env-updater/core"
    "time"
    "github.com/joho/godotenv"
    "strings"
)

// LoadEnv loads environment variables from a .env file
func init() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found. Proceeding with system environment variables...")
    }
}

// This function should map filenames or parts of filenames to Azure DevOps projects
func getProjectForFile(filename string) string {
    // projectfile mapping:
    projectMap := map[string]string{
        "frontend_": "gamepride-frontend",
        "api_":      "gamepride-api",
        "admin_":    "gamepride-admin",
    }

    for prefix, project := range projectMap {
        if strings.HasPrefix(filename, prefix) {
            return project
        }
    }
    // Default project if no match is found
    return "DefaultProject"
}

// Fetches secure file ID to be used in setting permissions
func getSecureFileId(ctx context.Context, pat, org, project, fileName string) (string, error) {
    apiURL := fmt.Sprintf(
        "https://dev.azure.com/%s/%s/_apis/distributedtask/securefiles?api-version=7.0-preview",
        org, project,
    )

    req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
    if err != nil {
        return "", fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var result struct {
        Value []struct {
            Id   string `json:"id"`
            Name string `json:"name"`
        } `json:"value"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    for _, file := range result.Value {
        if file.Name == fileName {
            return file.Id, nil
        }
    }

    return "", fmt.Errorf("secure file not found: %s", fileName)
}

// setSecureFilePermissions sets permissions for a pipeline on a secure file
func setSecureFilePermissions(ctx context.Context, pat, org, project, secureFileName string, pipelineId int) error {
    fileId, err := getSecureFileId(ctx, pat, org, project, secureFileName)
    if err != nil {
        return fmt.Errorf("failed to get secure file ID: %v", err)
    }

    apiURL := fmt.Sprintf(
        "https://dev.azure.com/%s/%s/_apis/pipelines/pipelinePermissions/securefile/%s?api-version=7.0-preview",
        org, project, fileId,
    )

    jsonPayload := map[string]interface{}{
        "allPipelines": map[string]interface{}{
            "authorized": false,
            "authorizedBy": nil,
            "authorizedOn": nil,
        },
        "pipelines": []map[string]interface{}{
            {
                "id": pipelineId,
                "authorized": true,
            },
        },
    }

    payloadBytes, err := json.Marshal(jsonPayload)
    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %v", err)
    }

    req, err := http.NewRequestWithContext(ctx, "PATCH", apiURL, bytes.NewReader(payloadBytes))
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }

    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed with status %d: %s", resp.StatusCode, string(bodyBytes))
    }

    return nil
}

// triggerCIByMatchablePart searches for a pipeline with the most matching letters in the string after the last dot of the filename
func triggerCIByMatchablePart(ctx context.Context, matchPart, project, filename string) error {
    pat := os.Getenv("AZURE_DEVOPS_PAT")
    org := os.Getenv("AZURE_DEVOPS_ORG")

    if pat == "" || org == "" || project == "" {
        return fmt.Errorf("missing environment variables: AZURE_DEVOPS_PAT, AZURE_DEVOPS_ORG, or project")
    }

    // Fetch all pipelines to find a match
    pipelinesURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/pipelines?api-version=7.1-preview.1", org, project)

    req, err := http.NewRequestWithContext(ctx, "GET", pipelinesURL, nil)
    if err != nil {
        return fmt.Errorf("failed to create request for pipelines: %v", err)
    }
    req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))

    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to fetch pipelines: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("failed to get pipelines: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
    }

    var pipelineList struct {
        Value []struct {
            Id   int    `json:"id"`
            Name string `json:"name"`
        } `json:"value"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&pipelineList); err != nil {
        return fmt.Errorf("failed to decode pipeline list: %v", err)
    }

    // Find the pipeline with the most matching letters for the part after the last dot
    var bestMatch struct {
        Pipeline struct {
            Id   int
            Name string
        }
        Score int
    }
    bestMatch.Score = -1 // Initialize with a score lower than possible

    for _, pipeline := range pipelineList.Value {
        score := calculateMatchScore(strings.ToLower(matchPart), strings.ToLower(pipeline.Name))
        if score > bestMatch.Score {
            bestMatch.Score = score
            bestMatch.Pipeline.Id = pipeline.Id
            bestMatch.Pipeline.Name = pipeline.Name
        }
    }

    if bestMatch.Score > 0 {
        // Set permissions for the pipeline on the secure file before triggering
        if err := setSecureFilePermissions(ctx, pat, org, project, filename, bestMatch.Pipeline.Id); err != nil {
            return fmt.Errorf("failed to set permissions for pipeline %d on file %s: %v", bestMatch.Pipeline.Id, filename, err)
        }

        // Trigger the best matching pipeline
        triggerURL := fmt.Sprintf("https://dev.azure.com/%s/%s/_apis/pipelines/%d/runs?api-version=7.1-preview.1", org, project, bestMatch.Pipeline.Id)

        jsonPayload := map[string]interface{}{
            "resources": map[string]interface{}{
                "repositories": map[string]interface{}{
                    
                },
            },
        }

        payloadBytes, err := json.Marshal(jsonPayload)
        if err != nil {
            return fmt.Errorf("failed to marshal JSON payload for CI trigger: %v", err)
        }

        req, err := http.NewRequestWithContext(ctx, "POST", triggerURL, bytes.NewReader(payloadBytes))
        if err != nil {
            return fmt.Errorf("failed to create CI trigger request: %v", err)
        }
        req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(":"+pat)))
        req.Header.Set("Content-Type", "application/json")

        resp, err := client.Do(req)
        if err != nil {
            return fmt.Errorf("failed to trigger CI/CD: %v", err)
        }
        defer resp.Body.Close()

        bodyBytes, _ := io.ReadAll(resp.Body)
        if resp.StatusCode != http.StatusCreated {
            if resp.StatusCode == http.StatusOK {
                // Consider logging this for debugging or monitoring
                // log.Printf("CI/CD trigger responded with status code 200")
            } else {
                return fmt.Errorf("CI/CD trigger failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
            }
        }

        log.Printf("Successfully triggered pipeline %s", bestMatch.Pipeline.Name)
        return nil
    } else {
        log.Printf("No matching pipeline found for matchable part %s", matchPart)
    }
    return nil // No matching pipeline found, but this isn't necessarily an error
}

func ProcessWebhookEvent(webhookData map[string]interface{}) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    repo, ok := webhookData["repository"].(map[string]interface{})
    if !ok {
        return fmt.Errorf("no repository data found")
    }

    fullName, ok := repo["full_name"].(string)
    if !ok {
        return fmt.Errorf("could not extract repository full name")
    }

    files, ok := webhookData["commits"].([]interface{})
    if !ok || len(files) == 0 {
        return fmt.Errorf("no files found in webhook")
    }

    for _, commitInterface := range files {
        commit, ok := commitInterface.(map[string]interface{})
        if !ok {
            log.Printf("Error processing commit data: %v", commitInterface)
            continue
        }

        modifiedFiles, ok := commit["modified"].([]interface{})
        if !ok {
            log.Printf("Error getting modified files from commit: %v", commit)
            continue
        }

        for _, fileInterface := range modifiedFiles {
            filename, ok := fileInterface.(string)
            if !ok {
                log.Printf("Error converting modified file to string: %v", fileInterface)
                continue
            }

            fileContent, err := core.FetchFileFromGitHub(fullName, filename)
            if err != nil {
                log.Printf("GitHub file fetch error for %s: %v", filename, err)
                continue
            }

            if err := os.WriteFile("security.txt", fileContent, 0644); err != nil {
                log.Printf("Failed to write temporary security.txt: %v", err)
                continue
            }

            project := getProjectForFile(filepath.Base(filename))
            if err := os.Setenv("AZURE_DEVOPS_PROJECT", project); err != nil {
                log.Printf("Failed to set environment variable for project: %v", err)
                continue
            }

            err = core.UpdateAzureDevOpsFile(ctx, filepath.Base(filename))
            if err != nil {
                if !isSuccessError(err) {
                    log.Printf("Azure DevOps update error for %s: %v", filename, err)
                    continue
                }
                log.Printf("File %s successfully updated in Azure DevOps", filename)
            } else {
                log.Printf("Successfully processed file: %s in project %s", filename, project)
            }

            // Trigger CI/CD based on the part of filename after last dot
            matchPart := getMatchablePartFromFilename(filepath.Base(filename))
            if err := triggerCIByMatchablePart(ctx, matchPart, project, filepath.Base(filename)); err != nil {
                log.Printf("Failed to trigger CI/CD for matchable part %s: %v", matchPart, err)
            }
        }

        // Delete the temporary security.txt file after processing all files in the commit
        if err := os.Remove("security.txt"); err != nil {
            log.Printf("Failed to delete security.txt: %v", err)
        } 
    }

    return nil
}

