package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// MavenGet handles Maven artifact retrieval
func MavenGet(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Maven get not yet implemented"})
	}
}

// MavenPut handles Maven artifact upload
func MavenPut(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Maven put not yet implemented"})
	}
}

// MavenHead checks Maven artifact existence
func MavenHead(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Maven head not yet implemented"})
	}
}
