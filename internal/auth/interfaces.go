package auth

// AuthInterface defines the interface for authentication operations
type AuthInterface interface {
	ValidateToken(tokenString string) (*Claims, error)
	CheckPermission(claims *Claims, repository string, permission Permission) bool
}

// Ensure AuthService implements AuthInterface
var _ AuthInterface = (*AuthService)(nil)
