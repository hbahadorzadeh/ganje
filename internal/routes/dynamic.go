package routes

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/controllers"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// RegisterDynamicRoutes registers dynamic artifact routes based on repository configurations
func RegisterDynamicRoutes(
	r *gin.Engine,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	cfg *config.Config,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	// Docker registry API version endpoint
	r.GET("/v2/", controllers.DockerAPIVersion())

	// Register routes for all existing repositories
	repos, err := db.ListRepositories(context.Background())
	if err != nil {
		// Log error but don't fail server startup
		return
	}

	for _, repo := range repos {
		registerRepositoryRoutes(r, repo, db, storageService, authService, messagingService, metricsService, cfg, authMiddleware, requireRead, requireWrite)
	}
}

// registerRepositoryRoutes registers routes for a specific repository
func registerRepositoryRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	cfg *config.Config,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	// Register artifact-specific routes based on repository type
	switch repo.ArtifactType {
	case "docker":
		registerDockerRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "maven":
		registerMavenRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "npm":
		registerNpmRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "cargo":
		registerCargoRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "go":
		registerGoRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "helm":
		registerHelmRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	case "generic":
		registerGenericRoutes(r, repo, db, storageService, authService, messagingService, metricsService, authMiddleware, requireRead, requireWrite)
	}
}

// registerDockerRoutes registers Docker registry routes
func registerDockerRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/v2/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/tags/list", requireRead, controllers.DockerListTags(db, repo.Name))
	repoGroup.GET("/manifests/:reference", requireRead, controllers.DockerGetManifest(db, storageService, repo.Name))
	repoGroup.PUT("/manifests/:reference", requireWrite, controllers.DockerPutManifest(db, storageService, messagingService, repo.Name))
	repoGroup.HEAD("/manifests/:reference", requireRead, controllers.DockerHeadManifest(db, repo.Name))
	repoGroup.DELETE("/manifests/:reference", requireWrite, controllers.DockerDeleteManifest(db, storageService, repo.Name))
	repoGroup.GET("/blobs/:digest", requireRead, controllers.DockerGetBlob(db, storageService, repo.Name))
	repoGroup.HEAD("/blobs/:digest", requireRead, controllers.DockerHeadBlob(db, repo.Name))
	repoGroup.POST("/blobs/uploads/", requireWrite, controllers.DockerStartUpload(db, repo.Name))
	repoGroup.PUT("/blobs/uploads/:uuid", requireWrite, controllers.DockerCompleteUpload(db, storageService, messagingService, repo.Name))
	repoGroup.PATCH("/blobs/uploads/:uuid", requireWrite, controllers.DockerChunkedUpload(db, storageService, repo.Name))
	repoGroup.GET("/blobs/uploads/:uuid", requireRead, controllers.DockerGetUploadStatus(db, repo.Name))
	repoGroup.DELETE("/blobs/uploads/:uuid", requireWrite, controllers.DockerCancelUpload(db, repo.Name))
}

// registerMavenRoutes registers Maven repository routes
func registerMavenRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/*path", requireRead, controllers.MavenGet(db, storageService, repo.Name))
	repoGroup.PUT("/*path", requireWrite, controllers.MavenPut(db, storageService, messagingService, repo.Name))
	repoGroup.HEAD("/*path", requireRead, controllers.MavenHead(db, repo.Name))
}

// registerNpmRoutes registers NPM repository routes
func registerNpmRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/*package", requireRead, controllers.NpmGet(db, storageService, repo.Name))
	repoGroup.PUT("/*package", requireWrite, controllers.NpmPut(db, storageService, messagingService, repo.Name))
}

// registerCargoRoutes registers Cargo (Rust) repository routes
func registerCargoRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/api/v1/crates", requireRead, controllers.CargoSearchCrates(db, repo.Name))
	repoGroup.GET("/api/v1/crates/:name", requireRead, controllers.CargoGetCrate(db, storageService, repo.Name))
	repoGroup.PUT("/api/v1/crates/new", requireWrite, controllers.CargoPutCrate(db, storageService, messagingService, repo.Name))
	repoGroup.DELETE("/api/v1/crates/:name/:version/yank", requireWrite, controllers.CargoYankCrate(db, repo.Name))
	repoGroup.PUT("/api/v1/crates/:name/:version/unyank", requireWrite, controllers.CargoUnyankCrate(db, repo.Name))
	repoGroup.GET("/api/v1/crates/:name/versions", requireRead, controllers.CargoGetVersions(db, repo.Name))
	repoGroup.GET("/api/v1/crates/:name/:version/download", requireRead, controllers.CargoDownload(db, storageService, repo.Name))
}

// registerGoRoutes registers Go module repository routes
func registerGoRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/*module", requireRead, controllers.GoGet(db, storageService, repo.Name))
	repoGroup.POST("/*module", requireWrite, controllers.GoPut(db, storageService, messagingService, repo.Name))
}

// registerHelmRoutes registers Helm chart repository routes
func registerHelmRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/index.yaml", requireRead, controllers.HelmGetIndex(db, storageService, repo.Name))
	repoGroup.GET("/charts/:chart-:version.tgz", requireRead, controllers.HelmGetChart(db, storageService, repo.Name))
	repoGroup.POST("/api/charts", requireWrite, controllers.HelmPutChart(db, storageService, messagingService, repo.Name))
}

// registerGenericRoutes registers generic artifact repository routes
func registerGenericRoutes(
	r *gin.Engine,
	repo *database.Repository,
	db database.DatabaseInterface,
	storageService storage.Storage,
	authService auth.AuthInterface,
	messagingService messaging.Publisher,
	metricsService *metrics.MetricsService,
	authMiddleware gin.HandlerFunc,
	requireRead gin.HandlerFunc,
	requireWrite gin.HandlerFunc,
) {
	repoGroup := r.Group("/" + repo.Name)
	repoGroup.Use(authMiddleware)
	
	repoGroup.GET("/*path", requireRead, controllers.GenericGet(db, storageService, repo.Name))
	repoGroup.PUT("/*path", requireWrite, controllers.GenericPut(db, storageService, messagingService, repo.Name))
	repoGroup.HEAD("/*path", requireRead, controllers.GenericHead(db, repo.Name))
	repoGroup.DELETE("/*path", requireWrite, controllers.GenericDelete(db, storageService, repo.Name))
}
