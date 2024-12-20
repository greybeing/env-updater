package services

import (
    "context"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
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

// This function should map filenames or parts of filenames to Azure DevOps projects
func getProjectForFile(filename string) string {
    // projectfile mapping:
    projectMap := map[string]string{
        "frontend_": "gamepride-frontend",
        "api_":  "gamepride-api",
        "admin_":   "gamepride-admin",
    }

    for prefix, project := range projectMap {
        if strings.HasPrefix(filename, prefix) {
            return project
        }
    }
    // Default project if no match is found
    return "DefaultProject"
}

func ProcessWebhookEvent(webhookData map[string]interface{}) error {
    // Create a context with cancellation
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

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

            // Save the fetched content to 'security.txt' temporarily
            if err := os.WriteFile("security.txt", fileContent, 0644); err != nil {
                log.Printf("Failed to write temporary security.txt: %v", err)
                continue
            }

            // Dynamically set the project based on the filename
            project := getProjectForFile(filepath.Base(filename))
            if err := os.Setenv("AZURE_DEVOPS_PROJECT", project); err != nil {
                log.Printf("Failed to set environment variable for project: %v", err)
                continue
            }

            // Update file in Azure DevOps (includes delete and upload)
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
        }   
            // Delete the temporary security.txt file after successful use
            if err := os.Remove("security.txt"); err != nil {
                log.Printf("Failed to delete security.txt: %v", err)
            } else {
                log.Println("Successfully deleted security.txt")
            }
        
    }

    return nil
}

// Helper function to determine if an error is actually a success with status code 200
func isSuccessError(err error) bool {
    return err != nil && containsStatusCode200(err.Error())
}

// Helper function to check if the error message contains "status code 200"
func containsStatusCode200(msg string) bool {
    return strings.Contains(msg, "status code 200")
}