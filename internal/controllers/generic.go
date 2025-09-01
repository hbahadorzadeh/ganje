package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// GenericGet handles generic artifact retrieval
func GenericGet(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		
		// Get artifact from storage
		reader, err := storageService.Retrieve(c.Request.Context(), repoName+"/"+path)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Artifact not found"})
			return
		}
		defer reader.Close()

		// Stream the content
		c.DataFromReader(http.StatusOK, -1, "application/octet-stream", reader, nil)
	}
}

// GenericPut handles generic artifact upload
func GenericPut(db database.DatabaseInterface, storageService storage.Storage, messagingService messaging.Publisher, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		
		// Store artifact in storage
		err := storageService.Store(c.Request.Context(), repoName+"/"+path, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store artifact"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Artifact uploaded successfully"})
	}
}

// GenericHead checks generic artifact existence
func GenericHead(db database.DatabaseInterface, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		
		// Check if artifact exists in database
		_, err := db.GetArtifactByPath(c.Request.Context(), repoName, path)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusOK)
	}
}

// GenericDelete deletes generic artifact
func GenericDelete(db database.DatabaseInterface, storageService storage.Storage, repoName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Param("path")
		
		// Delete from storage
		err := storageService.Delete(c.Request.Context(), repoName+"/"+path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete artifact"})
			return
		}

		// Delete from database
		err = db.DeleteArtifactByPath(c.Request.Context(), repoName, path)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete artifact metadata"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Artifact deleted successfully"})
	}
}
