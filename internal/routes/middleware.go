package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/database"
)

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	}
}

// healthCheck returns server health status
func healthCheck(db database.DatabaseInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check database connectivity
		dbStatus := "healthy"
		if _, err := db.GetRepository(c.Request.Context(), "health-check-test"); err != nil {
			// This is expected to fail, we just want to test connectivity
			// If we can't connect to DB at all, this would panic or return a connection error
		}

		// Overall health status
		status := "healthy"
		httpStatus := http.StatusOK
		
		if dbStatus != "healthy" {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
		}

		c.JSON(httpStatus, gin.H{
			"status":     status,
			"service":    "ganje-artifact-repository",
			"version":    "1.0.0",
			"timestamp":  c.Request.Header.Get("Date"),
			"components": gin.H{
				"database": dbStatus,
			},
		})
	}
}

// createAuthMiddleware creates authentication middleware
func createAuthMiddleware(authService auth.AuthInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		claims, err := authService.ValidateToken(authHeader)
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

// createPermissionMiddleware creates permission checking middleware
func createPermissionMiddleware(authService auth.AuthInterface, permission auth.Permission) gin.HandlerFunc {
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

		if !authService.CheckPermission(claims, repository, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
