package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OIDCConfig represents OIDC configuration
type OIDCConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	IssuerURL    string
}

// TokenResponse represents OAuth2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

// UserInfo represents OIDC user information
type UserInfo struct {
	Sub               string   `json:"sub"`
	Name              string   `json:"name"`
	PreferredUsername string   `json:"preferred_username"`
	Email             string   `json:"email"`
	EmailVerified     bool     `json:"email_verified"`
	Groups            []string `json:"groups"`
}

// OIDCService handles OIDC authentication flow
type OIDCService struct {
	config     OIDCConfig
	httpClient *http.Client
}

// NewOIDCService creates a new OIDC service
func NewOIDCService(config OIDCConfig) *OIDCService {
	return &OIDCService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExchangeCodeForToken exchanges authorization code for access token
func (o *OIDCService) ExchangeCodeForToken(ctx context.Context, code string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/token", o.config.IssuerURL)
	
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", o.config.RedirectURI)
	data.Set("client_id", o.config.ClientID)
	data.Set("client_secret", o.config.ClientSecret)

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information using access token
func (o *OIDCService) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	userInfoURL := fmt.Sprintf("%s/userinfo", o.config.IssuerURL)

	req, err := http.NewRequestWithContext(ctx, "GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// CreateJWTFromUserInfo creates a JWT token from user information
func (o *OIDCService) CreateJWTFromUserInfo(userInfo *UserInfo, jwtSecret string) (string, error) {
	// Map Dex groups to Ganje realms
	realms := o.mapGroupsToRealms(userInfo.Groups)

	claims := &Claims{
		Username: userInfo.PreferredUsername,
		Email:    userInfo.Email,
		Realms:   realms,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userInfo.Sub,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "ganje",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// mapGroupsToRealms maps Dex groups to Ganje realms
func (o *OIDCService) mapGroupsToRealms(groups []string) []string {
	var realms []string
	
	for _, group := range groups {
		switch group {
		case "admins":
			realms = append(realms, "admins")
		case "developers":
			realms = append(realms, "developers")
		default:
			// Default to developers realm for unknown groups
			realms = append(realms, "developers")
		}
	}

	// If no groups, default to developers
	if len(realms) == 0 {
		realms = []string{"developers"}
	}

	return realms
}
