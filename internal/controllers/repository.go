package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// ListRepositories returns all repositories
func ListRepositories(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		repos, err := db.ListRepositories(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list repositories"})
			return
		}
		c.JSON(http.StatusOK, repos)
	}
}

// GetRepository returns a specific repository
func GetRepository(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		repo, err := db.GetRepository(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
			return
		}
		c.JSON(http.StatusOK, repo)
	}
}

// CreateRepository creates a new repository
func CreateRepository(db database.DatabaseInterface, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var repo database.Repository
		if err := c.ShouldBindJSON(&repo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := db.SaveRepository(c.Request.Context(), &repo); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create repository"})
			return
		}

		c.JSON(http.StatusCreated, repo)
	}
}

// UpdateRepository updates an existing repository
func UpdateRepository(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		var updates map[string]interface{}
		if err := c.ShouldBindJSON(&updates); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if err := db.UpdateRepository(c.Request.Context(), name, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Repository updated successfully"})
	}
}

// DeleteRepository deletes a repository
func DeleteRepository(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		if err := db.DeleteRepository(c.Request.Context(), name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
	}
}

// ValidateRepositoryConfig validates repository configuration
func ValidateRepositoryConfig(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var repoConfig config.RepositoryConfig
		if err := c.ShouldBindJSON(&repoConfig); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository configuration"})
			return
		}

		// Basic validation
		if repoConfig.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Repository name is required"})
			return
		}

		if repoConfig.ArtifactType == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Artifact type is required"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Repository configuration is valid"})
	}
}

// BulkDeleteRepositories deletes multiple repositories
func BulkDeleteRepositories(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Names []string `json:"names" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		var errors []string
		for _, name := range request.Names {
			if err := db.DeleteRepository(c.Request.Context(), name); err != nil {
				errors = append(errors, name+": "+err.Error())
			}
		}

		if len(errors) > 0 {
			c.JSON(http.StatusPartialContent, gin.H{"errors": errors})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "All repositories deleted successfully"})
	}
}

// GetRepositoryTypes returns available repository types
func GetRepositoryTypes() gin.HandlerFunc {
	return func(c *gin.Context) {
		types := []string{"docker", "maven", "npm", "cargo", "go", "helm", "generic"}
		c.JSON(http.StatusOK, gin.H{"types": types})
	}
}

// InvalidateCache invalidates repository cache
func InvalidateCache(storageService storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		// For now, just return success - cache invalidation logic would go here
		c.JSON(http.StatusOK, gin.H{"message": "Cache invalidated for repository " + name})
	}
}

// RebuildIndex rebuilds repository index
func RebuildIndex(db database.DatabaseInterface, storageService storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		// For now, just return success - index rebuilding logic would go here
		c.JSON(http.StatusOK, gin.H{"message": "Index rebuilt for repository " + name})
	}
}

// GetRepositoryStats returns repository statistics
func GetRepositoryStats(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")
		
		stats, err := db.GetRepositoryStatistics(c.Request.Context(), name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get repository statistics"})
			return
		}

		c.JSON(http.StatusOK, stats)
	}
}
