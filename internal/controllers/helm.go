package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// HelmGetIndex gets Helm chart index
func HelmGetIndex(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Helm get index not yet implemented"})
	}
}

// HelmGetChart gets Helm chart
func HelmGetChart(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Helm get chart not yet implemented"})
	}
}

// HelmPutChart uploads Helm chart
func HelmPutChart(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Helm put chart not yet implemented"})
	}
}
