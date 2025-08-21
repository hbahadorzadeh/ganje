package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims
type Claims struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Realms   []string `json:"realms"`
	jwt.RegisteredClaims
}

// Permission represents access permissions
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
	PermissionAdmin Permission = "admin"
)

// AuthService handles authentication and authorization
type AuthService struct {
	jwtSecret   string
	oauthServer string
	realms      map[string][]Permission
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret, oauthServer string, realms map[string][]Permission) *AuthService {
	return &AuthService{
		jwtSecret:   jwtSecret,
		oauthServer: oauthServer,
		realms:      realms,
	}
}

// ValidateToken validates JWT token and returns claims
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// CheckPermission checks if user has required permission for repository
func (a *AuthService) CheckPermission(claims *Claims, repository string, permission Permission) bool {
	// Check each realm the user belongs to
	for _, userRealm := range claims.Realms {
		if permissions, exists := a.realms[userRealm]; exists {
			for _, perm := range permissions {
				if perm == permission || perm == PermissionAdmin {
					return true
				}
			}
		}
	}

	return false
}

// AuthContext represents authentication context
type AuthContext struct {
	Username string
	Email    string
	Realms   []string
}

// contextKey is used for context values
type contextKey string

const authContextKey contextKey = "auth_context"

// WithAuthContext adds authentication context to context
func WithAuthContext(ctx context.Context, authCtx *AuthContext) context.Context {
	return context.WithValue(ctx, authContextKey, authCtx)
}

// GetAuthContext retrieves authentication context from context
func GetAuthContext(ctx context.Context) (*AuthContext, bool) {
	authCtx, ok := ctx.Value(authContextKey).(*AuthContext)
	return authCtx, ok
}

// Middleware represents authentication middleware interface
type Middleware interface {
	Authenticate(next func(ctx context.Context) error) func(ctx context.Context) error
	RequirePermission(permission Permission, repository string) func(next func(ctx context.Context) error) func(ctx context.Context) error
}
