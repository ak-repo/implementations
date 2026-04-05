// Package services contains business logic and external service integrations.
// The oauth package handles Google OAuth 2.0 authentication flow.
package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"oauth-demo/internal/models"
)

// OAuthService handles Google OAuth 2.0 authentication.
type OAuthService struct {
	config     *oauth2.Config
	stateStore map[string]bool // In production, use Redis or encrypted cookies
}

// NewOAuthService creates a new OAuth service with Google configuration.
// Reads credentials from environment variables.
func NewOAuthService() (*OAuthService, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return nil, fmt.Errorf("missing required OAuth environment variables")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &OAuthService{
		config:     config,
		stateStore: make(map[string]bool),
	}, nil
}

// GenerateState creates a random state parameter for CSRF protection.
// The state is a cryptographically secure random string that must be
// verified when the user returns from Google's OAuth flow.
func (s *OAuthService) GenerateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	s.stateStore[state] = true
	return state
}

// VerifyState checks if the state parameter is valid.
// This prevents CSRF attacks by ensuring the state matches what we generated.
func (s *OAuthService) VerifyState(state string) bool {
	if s.stateStore[state] {
		delete(s.stateStore, state) // Use once and delete
		return true
	}
	return false
}

// GetAuthURL returns the Google OAuth consent screen URL.
// The user should be redirected to this URL to begin the OAuth flow.
func (s *OAuthService) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeCode exchanges the authorization code for an access token.
// This is called after Google redirects the user back to our callback URL.
func (s *OAuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo fetches the user's profile from Google's UserInfo endpoint.
// Requires a valid OAuth2 token obtained from ExchangeCode.
func (s *OAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*models.GoogleUserInfo, error) {
	// Create an HTTP client with the OAuth token
	client := s.config.Client(ctx, token)

	// Google's OpenID Connect UserInfo endpoint
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo models.GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}
