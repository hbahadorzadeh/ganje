package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// DockerAPIVersion returns Docker registry API version
func DockerAPIVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Docker-Distribution-API-Version", "registry/2.0")
		c.JSON(http.StatusOK, gin.H{})
	}
}

// DockerListTags lists Docker image tags
func DockerListTags(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker list tags not yet implemented"})
	}
}

// DockerGetManifest gets Docker image manifest
func DockerGetManifest(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker get manifest not yet implemented"})
	}
}

// DockerPutManifest puts Docker image manifest
func DockerPutManifest(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker put manifest not yet implemented"})
	}
}

// DockerHeadManifest checks Docker image manifest existence
func DockerHeadManifest(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker head manifest not yet implemented"})
	}
}

// DockerDeleteManifest deletes Docker image manifest
func DockerDeleteManifest(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker delete manifest not yet implemented"})
	}
}

// DockerGetBlob gets Docker image blob
func DockerGetBlob(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker get blob not yet implemented"})
	}
}

// DockerHeadBlob checks Docker image blob existence
func DockerHeadBlob(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker head blob not yet implemented"})
	}
}

// DockerStartUpload starts Docker blob upload
func DockerStartUpload(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker start upload not yet implemented"})
	}
}

// DockerCompleteUpload completes Docker blob upload
func DockerCompleteUpload(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker complete upload not yet implemented"})
	}
}

// DockerChunkedUpload handles Docker chunked upload
func DockerChunkedUpload(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker chunked upload not yet implemented"})
	}
}

// DockerGetUploadStatus gets Docker upload status
func DockerGetUploadStatus(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker get upload status not yet implemented"})
	}
}

// DockerCancelUpload cancels Docker upload
func DockerCancelUpload(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Docker cancel upload not yet implemented"})
	}
}
