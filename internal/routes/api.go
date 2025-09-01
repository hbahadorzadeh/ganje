package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/controllers"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// RegisterAPIRoutes registers all API routes
func RegisterAPIRoutes(
	r *gin.Engine,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	oidcService *auth.OIDCService,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	cfg *config.Config,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
	requireAdmin gin.HandlerFunc,
) {
	api := r.Group("/api/v1")

	// Authentication endpoints
	if oidcService != nil {
		api.POST("/auth/callback", controllers.HandleAuthCallback(oidcService, cfg))
	}

	// Repository management
	api.GET("/repositories", authMiddleware, controllers.ListRepositories(db))
	api.GET("/repositories/:name", authMiddleware, controllers.GetRepository(db))
	api.POST("/repositories", authMiddleware, requireAdmin, controllers.CreateRepository(db, cfg))
	api.PUT("/repositories/:name", authMiddleware, requireAdmin, controllers.UpdateRepository(db))
	api.DELETE("/repositories/:name", authMiddleware, requireAdmin, controllers.DeleteRepository(db))
	
	// Repository validation and bulk operations
	api.POST("/repositories/validate", authMiddleware, requireAdmin, controllers.ValidateRepositoryConfig(cfg))
	api.DELETE("/repositories", authMiddleware, requireAdmin, controllers.BulkDeleteRepositories(db))
	api.GET("/repository-types", authMiddleware, controllers.GetRepositoryTypes())

	// Cache management
	api.DELETE("/repositories/:name/cache", authMiddleware, requireAdmin, controllers.InvalidateCache(storageService))
	api.POST("/repositories/:name/reindex", authMiddleware, requireAdmin, controllers.RebuildIndex(db, storageService))

	// Statistics
	api.GET("/repositories/:name/stats", authMiddleware, controllers.GetRepositoryStats(db))

	// Artifacts (admin portal)
	api.GET("/repositories/:name/artifacts", authMiddleware, requireRead, controllers.ListArtifacts(db))
	api.GET("/repositories/:name/artifacts/stats", authMiddleware, requireRead, controllers.GetArtifactStats(db))
	api.POST("/repositories/:name/artifacts", authMiddleware, requireWrite, controllers.UploadArtifact(db, storageService, messagingService))
	api.POST("/repositories/:name/artifacts/move", authMiddleware, requireWrite, controllers.MoveArtifact(db, storageService))
	api.POST("/repositories/:name/artifacts/copy", authMiddleware, requireWrite, controllers.CopyArtifact(db, storageService))

	// Webhooks (admin only)
	api.GET("/repositories/:name/webhooks", authMiddleware, requireAdmin, controllers.ListWebhooks(db))
	api.POST("/repositories/:name/webhooks", authMiddleware, requireAdmin, controllers.CreateWebhook(db))
	api.GET("/repositories/:name/webhooks/:id", authMiddleware, requireAdmin, controllers.GetWebhook(db))
	api.PUT("/repositories/:name/webhooks/:id", authMiddleware, requireAdmin, controllers.UpdateWebhook(db))
	api.DELETE("/repositories/:name/webhooks/:id", authMiddleware, requireAdmin, controllers.DeleteWebhook(db))

	// Search
	api.GET("/search", authMiddleware, requireRead, controllers.SearchArtifacts(db))
}
