package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
)

// ListWebhooks returns webhooks for a repository
func ListWebhooks(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		webhooks, err := db.ListWebhooksByRepository(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list webhooks"})
			return
		}

		c.JSON(http.StatusOK, webhooks)
	}
}

// CreateWebhook creates a new webhook
func CreateWebhook(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		var webhook database.Webhook
		if err := c.ShouldBindJSON(&webhook); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := db.CreateWebhook(c.Request.Context(), name, &webhook); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
			return
		}

		c.JSON(http.StatusCreated, webhook)
	}
}

// GetWebhook returns a specific webhook
func GetWebhook(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			return
		}

		webhook, err := db.GetWebhook(c.Request.Context(), uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
			return
		}

		c.JSON(http.StatusOK, webhook)
	}
}

// UpdateWebhook updates an existing webhook
func UpdateWebhook(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			return
		}

		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := db.UpdateWebhook(c.Request.Context(), uint(id), updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update webhook"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Webhook updated successfully"})
	}
}

// DeleteWebhook deletes a webhook
func DeleteWebhook(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
			return
		}

		if err := db.DeleteWebhook(c.Request.Context(), uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
	}
}
