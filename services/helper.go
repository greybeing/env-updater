package services

import (
    "strings"
)

// Helper function that Extracts the part of the filename after the last dot for matching
func getMatchablePartFromFilename(filename string) string {
    parts := strings.Split(filename, ".")
    if len(parts) > 1 {
        return parts[len(parts)-1] // Return the last part after the last dot
    }
    return filename // If no dot, return the whole filename
}

// Helper function to calculateMatchScore returns a score based on how many characters match between two strings, non-sequentially
func calculateMatchScore(str1, str2 string) int {
    score := 0
    for _, char := range str1 {
        if strings.ContainsRune(str2, char) {
            score++
        }
    }
    return score
}

// Helper function to determine if an error is actually a success with status code 200
func isSuccessError(err error) bool {
    return err != nil && containsStatusCode200(err.Error())
}

// Helper function to check if the error message contains "status code 200"
func containsStatusCode200(msg string) bool {
    return strings.Contains(msg, "status code 200")
}