package core

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
    "github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
	"github.com/joho/godotenv"
)



// LoadEnv loads environment variables from a .env file
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Proceeding with system environment variables...")
	}
}


func VerifyWebhookSignature(payload []byte, signature string) bool {
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret == "" {
		log.Println("GITHUB_WEBHOOK_SECRET is not set")
		return false
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	expectedSignature := "sha256=" + hex.EncodeToString(h.Sum(nil))

	signature = strings.TrimSpace(signature) // Normalize input
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func FetchFileFromGitHub(repoFullName, filePath string) ([]byte, error) {
	// Get GitHub token
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GitHub token not set")
	}

	// Create OAuth2 client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	// Create GitHub client
	client := github.NewClient(tc)

	// Parse repository owner and name
	owner, repo, err := SplitRepositoryFullName(repoFullName)
	if err != nil {
		return nil, fmt.Errorf("invalid repository name: %v", err)
	}

	// Fetch branch/ref
	ref := os.Getenv("GITHUB_REF")
	if ref == "" {
		ref = "main"
	}

	// Get file content
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		filePath,
		&github.RepositoryContentGetOptions{Ref: ref},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get file content: %v", err)
	}

	// Ensure fileContent is not nil
	if fileContent == nil {
		return nil, fmt.Errorf("file not found at path: %s", filePath)
	}

	// Decode file content
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode file content: %v", err)
	}

	return []byte(content), nil
}

// SplitRepositoryFullName splits a full repository name into owner and repo.
// It expects the format "owner/repo".
func SplitRepositoryFullName(repoFullName string) (string, string, error) {
	parts := strings.Split(repoFullName, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository full name: %s, expected format 'owner/repo'", repoFullName)
	}
	return parts[0], parts[1], nil
}
