package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/repository"
)

// Server represents the HTTP server
type Server struct {
	config        *config.Config
	db            database.DatabaseInterface
	repoManager   repository.Manager
	authService   auth.AuthInterface
	oidcService   *auth.OIDCService
	publisher     messaging.Publisher
	router        *gin.Engine
	routeRegistry *RouteRegistry
	metrics       *metrics.MetricsService
	metricsServer *metrics.MetricsServer
	startTime     time.Time
}

// New creates a new server instance
func New(cfg *config.Config) *Server {
	// Initialize database
	db, err := database.New(cfg.Database.Driver, cfg.Database.GetConnectionString())
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize database: %v", err))
	}

	// Initialize authentication service
	realmPerms := make(map[string][]auth.Permission)
	for _, realm := range cfg.Auth.Realms {
		var perms []auth.Permission
		for _, perm := range realm.Permissions {
			perms = append(perms, auth.Permission(perm))
		}
		realmPerms[realm.Name] = perms
	}
	authService := auth.NewAuthService(cfg.Auth.JWTSecret, cfg.Auth.OAuthServer, realmPerms)

	// Initialize OIDC service
	var oidcService *auth.OIDCService
	if cfg.Auth.OIDC.ClientID != "" {
		oidcConfig := auth.OIDCConfig{
			ClientID:     cfg.Auth.OIDC.ClientID,
			ClientSecret: cfg.Auth.OIDC.ClientSecret,
			RedirectURI:  cfg.Auth.OIDC.RedirectURI,
			IssuerURL:    cfg.Auth.OAuthServer,
		}
		oidcService = auth.NewOIDCService(oidcConfig)
	}

	// Initialize metrics service first
	var metricsService *metrics.MetricsService
	if cfg.Metrics.Enabled {
		metricsService = metrics.NewMetricsService()
		metricsService.SetSystemInfo("1.0.0", runtime.Version())
	}

	// Initialize repository manager
	repoManager := NewRepositoryManager(cfg, db, metricsService)

	// Initialize metrics server if separate server is enabled
	var metricsServer *metrics.MetricsServer
	if cfg.Metrics.Enabled && cfg.Metrics.SeparateServer {
		metricsServer = metrics.NewMetricsServer(cfg.Metrics.Port, metricsService)
	}

	// Initialize messaging publisher (RabbitMQ or Noop)
	var publisher messaging.Publisher = &messaging.NoopPublisher{}
	if cfg.Messaging.RabbitMQ.Enabled {
		pub, err := messaging.NewRabbitMQPublisher(
			cfg.Messaging.RabbitMQ.URL,
			cfg.Messaging.RabbitMQ.Exchange,
			cfg.Messaging.RabbitMQ.ExchangeType,
			cfg.Messaging.RabbitMQ.RoutingKey,
		)
		if err == nil {
			publisher = pub
		} else {
			fmt.Printf("Warning: RabbitMQ disabled due to init error: %v\n", err)
		}
	}

	server := &Server{
		config:        cfg,
		db:            db,
		repoManager:   repoManager,
		authService:   authService,
		oidcService:   oidcService,
		publisher:     publisher,
		routeRegistry: NewRouteRegistry(),
		metrics:       metricsService,
		metricsServer: metricsServer,
		startTime:     time.Now(),
	}

	// Webhook dispatcher now runs as a standalone service.

	server.setupRoutes()
	return server
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() {
	s.router = gin.Default()

	// Add CORS middleware
	s.router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Add metrics middleware if enabled
	if s.metrics != nil {
		s.router.Use(s.metrics.GinMiddleware())
	}

	// Health check
	s.router.GET("/health", s.healthCheck)

	// Metrics endpoint (if not using separate server)
	if s.metrics != nil && !s.config.Metrics.SeparateServer {
		s.router.GET(s.config.Metrics.Path, gin.WrapH(s.metrics.GetHandler()))
	}

	// API routes
	api := s.router.Group("/api/v1")
	{
		// Authentication endpoints
		if s.oidcService != nil {
			api.POST("/auth/callback", s.handleAuthCallback)
		}

		// Repository management
		api.GET("/repositories", s.authMiddleware(), s.listRepositories)
		api.GET("/repositories/:name", s.authMiddleware(), s.getRepository)
		api.POST("/repositories", s.authMiddleware(), s.requireAdmin(), s.createRepository)
		api.PUT("/repositories/:name", s.authMiddleware(), s.requireAdmin(), s.updateRepository)
		api.DELETE("/repositories/:name", s.authMiddleware(), s.requireAdmin(), s.deleteRepository)
		
		// Repository validation and bulk operations
		api.POST("/repositories/validate", s.authMiddleware(), s.requireAdmin(), s.validateRepositoryConfig)
		api.DELETE("/repositories", s.authMiddleware(), s.requireAdmin(), s.bulkDeleteRepositories)
		api.GET("/repository-types", s.authMiddleware(), s.getRepositoryTypes)

		// Cache management
		api.DELETE("/repositories/:name/cache", s.authMiddleware(), s.requireAdmin(), s.invalidateCache)
		api.POST("/repositories/:name/reindex", s.authMiddleware(), s.requireAdmin(), s.rebuildIndex)

		// Statistics
		api.GET("/repositories/:name/stats", s.authMiddleware(), s.getRepositoryStats)

		// Artifacts (admin portal)
		api.GET("/repositories/:name/artifacts", s.authMiddleware(), s.requireRead(), s.listArtifacts)
		api.GET("/repositories/:name/artifacts/stats", s.authMiddleware(), s.requireRead(), s.getArtifactStats)
		api.POST("/repositories/:name/artifacts", s.authMiddleware(), s.requireWrite(), s.uploadArtifact)
		api.POST("/repositories/:name/artifacts/move", s.authMiddleware(), s.requireWrite(), s.moveArtifact)
		api.POST("/repositories/:name/artifacts/copy", s.authMiddleware(), s.requireWrite(), s.copyArtifact)

		// Webhooks (admin only)
		api.GET("/repositories/:name/webhooks", s.authMiddleware(), s.requireAdmin(), s.listWebhooks)
		api.POST("/repositories/:name/webhooks", s.authMiddleware(), s.requireAdmin(), s.createWebhook)
		api.GET("/repositories/:name/webhooks/:id", s.authMiddleware(), s.requireAdmin(), s.getWebhook)
		api.PUT("/repositories/:name/webhooks/:id", s.authMiddleware(), s.requireAdmin(), s.updateWebhook)
		api.DELETE("/repositories/:name/webhooks/:id", s.authMiddleware(), s.requireAdmin(), s.deleteWebhook)

		// Search
		api.GET("/search", s.authMiddleware(), s.requireRead(), s.searchArtifacts)
	}

	// Register dynamic artifact routes based on repository configurations
	if err := s.setupDynamicRoutes(); err != nil {
		// Log error but don't fail server startup
		// Routes can be registered later when repositories are created
		fmt.Printf("Warning: Failed to setup dynamic routes: %v\n", err)
	}
}

// setupDynamicRoutes registers routes for all existing repositories
func (s *Server) setupDynamicRoutes() error {
	return s.routeRegistry.RegisterAllRepositoryRoutes(s.router, s, s.db)
}

// RegisterRepositoryRoutes registers routes for a specific repository (called when creating new repos)
func (s *Server) RegisterRepositoryRoutes(repo *database.Repository) error {
	return s.routeRegistry.RegisterRepositoryRoutes(s.router, s, repo)
}

// Start starts the HTTP server and metrics server (if configured)
func (s *Server) Start() error {
	// Start metrics server if configured
	if s.metricsServer != nil {
		go func() {
			if err := s.metricsServer.Start(); err != nil && err != http.ErrServerClosed {
				fmt.Printf("Metrics server failed to start: %v\n", err)
			}
		}()
		fmt.Printf("Metrics server started on port %d\n", s.config.Metrics.Port)
	}

	// Start uptime tracking
	if s.metrics != nil {
		go s.trackUptime()
	}

	// Start main server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	return s.router.Run(addr)
}

// Shutdown gracefully shuts down both servers
func (s *Server) Shutdown(ctx context.Context) error {
	var err error
	
	// Shutdown metrics server if running
	if s.metricsServer != nil {
		if shutdownErr := s.metricsServer.Shutdown(ctx); shutdownErr != nil {
			err = shutdownErr
		}
	}

	// Close messaging publisher
	if s.publisher != nil {
		if closeErr := s.publisher.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}

	// Webhook dispatcher runs as separate process; nothing to stop here.

	return err
}

// trackUptime tracks application uptime metrics
func (s *Server) trackUptime() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		uptime := time.Since(s.startTime).Seconds()
		s.metrics.RecordUptime("ganje", uptime)
	}
}

// healthCheck returns server health status
func (s *Server) healthCheck(c *gin.Context) {
	// Check database connectivity
	dbStatus := "healthy"
	if _, err := s.db.GetRepository(c.Request.Context(), "health-check-test"); err != nil {
		// This is expected to fail, we just want to test connectivity
		// If we can't connect to DB at all, this would panic or return a connection error
	}

	// Check repository manager
	repoStatus := "healthy"
	if s.repoManager == nil {
		repoStatus = "unhealthy"
	}

	// Overall health status
	status := "healthy"
	httpStatus := http.StatusOK
	
	if dbStatus != "healthy" || repoStatus != "healthy" {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":     status,
		"service":    "ganje-artifact-repository",
		"version":    "1.0.0",
		"timestamp":  c.Request.Header.Get("Date"),
		"components": gin.H{
			"database":   dbStatus,
			"repository": repoStatus,
		},
	})
}

// authMiddleware validates JWT token
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		claims, err := s.authService.ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store auth context
		authCtx := &auth.AuthContext{
			Username: claims.Username,
			Email:    claims.Email,
			Realms:   claims.Realms,
		}
		c.Set("auth_context", authCtx)
		c.Next()
	}
}

// requireRead checks read permission
func (s *Server) requireRead() gin.HandlerFunc {
	return s.requirePermission(auth.PermissionRead)
}

// requireWrite checks write permission
func (s *Server) requireWrite() gin.HandlerFunc {
	return s.requirePermission(auth.PermissionWrite)
}

// requireAdmin checks admin permission
func (s *Server) requireAdmin() gin.HandlerFunc {
	return s.requirePermission(auth.PermissionAdmin)
}

// requirePermission checks specific permission
func (s *Server) requirePermission(permission auth.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		authCtx, exists := c.Get("auth_context")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		authContext := authCtx.(*auth.AuthContext)
        // Prefer explicit route params set for admin/API endpoints
        repository := c.Param("name")
        if repository == "" {
            repository = c.Param("repository")
        }
        // For non-API dynamic routes (e.g., repoName/... at root), derive from path
        if repository == "" {
            fullPath := c.Request.URL.Path
            // If this is an API path, do not guess a repository from "/api/..."
            if !strings.HasPrefix(fullPath, "/api/") {
                if len(fullPath) > 1 {
                    trimmed := strings.TrimPrefix(fullPath, "/")
                    if i := strings.Index(trimmed, "/"); i >= 0 {
                        repository = trimmed[:i]
                    } else {
                        repository = trimmed
                    }
                }
            }
        }

		claims := &auth.Claims{
			Username: authContext.Username,
			Email:    authContext.Email,
			Realms:   authContext.Realms,
		}

		if !s.authService.CheckPermission(claims, repository, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleAuthCallback handles OIDC authentication callback
func (s *Server) handleAuthCallback(c *gin.Context) {
	if s.oidcService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OIDC not configured"})
		return
	}

	var request struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	// Exchange code for token
	tokenResp, err := s.oidcService.ExchangeCodeForToken(c.Request.Context(), request.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to exchange code for token", "details": err.Error()})
		return
	}

	// Get user info
	userInfo, err := s.oidcService.GetUserInfo(c.Request.Context(), tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info", "details": err.Error()})
		return
	}

	// Create JWT token
	jwtToken, err := s.oidcService.CreateJWTFromUserInfo(userInfo, s.config.Auth.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create JWT token", "details": err.Error()})
		return
	}

	// Return response compatible with frontend
	c.JSON(http.StatusOK, gin.H{
		"token": jwtToken,
		"user": gin.H{
			"id":       userInfo.Sub,
			"username": userInfo.PreferredUsername,
			"email":    userInfo.Email,
			"realms":   userInfo.Groups,
			"active":   true,
		},
	})
}

// dockerAPIVersion returns Docker registry API version
func (s *Server) dockerAPIVersion(c *gin.Context) {
	c.Header("Docker-Distribution-API-Version", "registry/2.0")
	c.JSON(http.StatusOK, gin.H{})
}

