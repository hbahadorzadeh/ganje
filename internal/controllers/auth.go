package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
)

// HandleAuthCallback handles OIDC authentication callback
func HandleAuthCallback(oidcService *auth.OIDCService, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if oidcService == nil {
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
		tokenResp, err := oidcService.ExchangeCodeForToken(c.Request.Context(), request.Code)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to exchange code for token", "details": err.Error()})
			return
		}

		// Get user info
		userInfo, err := oidcService.GetUserInfo(c.Request.Context(), tokenResp.AccessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user info", "details": err.Error()})
			return
		}

		// Create JWT token
		jwtToken, err := oidcService.CreateJWTFromUserInfo(userInfo, cfg.Auth.JWTSecret)
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
}
