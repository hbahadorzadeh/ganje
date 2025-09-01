package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// ListArtifacts returns artifacts for a repository
func ListArtifacts(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		artifacts, err := db.GetArtifactsByRepository(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list artifacts"})
			return
		}

		c.JSON(http.StatusOK, artifacts)
	}
}

// GetArtifactStats returns artifact statistics
func GetArtifactStats(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		stats, err := db.GetRepositoryStatistics(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get artifact statistics"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}

// UploadArtifact handles artifact upload
func UploadArtifact(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		// For now, return not implemented - full implementation would handle file upload
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Upload artifact not yet implemented for repository " + name})
	}
}

// MoveArtifact moves an artifact
func MoveArtifact(db database.DatabaseInterface, storageService storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		// For now, return not implemented
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Move artifact not yet implemented for repository " + name})
	}
}

// CopyArtifact copies an artifact
func CopyArtifact(db database.DatabaseInterface, storageService storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		// For now, return not implemented
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Copy artifact not yet implemented for repository " + name})
	}
}

// SearchArtifacts searches for artifacts
func SearchArtifacts(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Query("q")
		limitStr := c.DefaultQuery("limit", "10")
		
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			limit = 10
		}

		// For now, return empty results - full implementation would search artifacts
		c.JSON(http.StatusOK, gin.H{
			"query": query,
			"limit": limit,
			"results": []interface{}{},
		})
	}
}
