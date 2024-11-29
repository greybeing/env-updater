package core

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// FetchUpdatedFile fetches the specified file from GitHub
func FetchUpdatedFile(filePath, owner, repo, branch string) string {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		fmt.Println("GITHUB_TOKEN is not set")
		return ""
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, branch, filePath)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return ""
	}
	req.Header.Set("Authorization", "Bearer "+githubToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("Error fetching file: %v, Status: %d\n", err, resp.StatusCode)
		return ""
	}
	defer resp.Body.Close()

	outputPath := "/tmp/" + filePath
	os.MkdirAll("/tmp/env", os.ModePerm) // Create parent directory if needed
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return ""
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return ""
	}

	return outputPath
}
