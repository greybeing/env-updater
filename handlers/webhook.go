package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
    "github.com/gin-gonic/gin"
	"env-updater/core"
	"env-updater/services"
)

func HandleWebhook(c *gin.Context) {
	// Read webhook payload
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Payload read error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
		return
	}

	// Verify webhook signature
	if !core.VerifyWebhookSignature(payload, c.GetHeader("X-Hub-Signature-256")) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized webhook"})
		return
	}

	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		log.Printf("Payload parsing error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook format"})
		return
	}

	// Process webhook
	if err := services.ProcessWebhookEvent(webhookData); err != nil {
		log.Printf("Webhook processing error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Processing failed"})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"status": "Webhook processed successfully"})
}