package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// SetupRoutes configures all routes with dependency injection
func SetupRoutes(
	r *gin.Engine,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	oidcService *auth.OIDCService,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	cfg *config.Config,
) {
	// Add CORS middleware
	r.Use(corsMiddleware())

	// Add metrics middleware if enabled
	if metricsService != nil {
		r.Use(metricsService.GinMiddleware())
	}

	// Health check
	r.GET("/health", healthCheck(db))

	// Metrics endpoint (if not using separate server)
	if metricsService != nil && !cfg.Metrics.SeparateServer {
		r.GET(cfg.Metrics.Path, gin.WrapH(metricsService.GetHandler()))
	}

	// Authentication middleware factory
	authMiddleware := createAuthMiddleware(authService)
	requireRead := createPermissionMiddleware(authService, auth.PermissionRead)
	requireWrite := createPermissionMiddleware(authService, auth.PermissionWrite)
	requireAdmin := createPermissionMiddleware(authService, auth.PermissionAdmin)

	// Setup API routes
	RegisterAPIRoutes(r, db, storageService, authService, oidcService, messagingService, metricsService, cfg, authMiddleware, requireRead, requireWrite, requireAdmin)

	// Setup dynamic artifact routes
	RegisterDynamicRoutes(r, db, storageService, authService, messagingService, metricsService, cfg, authMiddleware, requireRead, requireWrite)
}
