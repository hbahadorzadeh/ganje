package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// GoGet handles Go module retrieval
func GoGet(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Go get not yet implemented"})
	}
}

// GoPut handles Go module upload
func GoPut(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Go put not yet implemented"})
	}
}
