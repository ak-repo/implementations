// Package handlers contains HTTP request handlers for the application.
package handlers

import (
	"encoding/json"
	"net/http"
	"oauth-demo/internal/middleware"
	"oauth-demo/internal/models"
	"oauth-demo/internal/repository"
	"oauth-demo/internal/services"
	"os"

	"github.com/go-chi/chi/v5"
)

// AuthHandler handles OAuth-related HTTP requests.
type AuthHandler struct {
	oauthService   *services.OAuthService
	jwtService     *services.JWTService
	userRepository repository.UserRepository
	frontendURL    string
}

// NewAuthHandler creates a new authentication handler.
func NewAuthHandler(
	oauthService *services.OAuthService,
	jwtService *services.JWTService,
	userRepository repository.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		oauthService:   oauthService,
		jwtService:     jwtService,
		userRepository: userRepository,
		frontendURL:    os.Getenv("FRONTEND_URL"),
	}
}

// RegisterRoutes registers all auth-related routes.
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/auth/google/login", h.GoogleLogin)
	r.Get("/auth/google/callback", h.GoogleCallback)
	r.Post("/auth/logout", h.Logout)
}

// GoogleLogin initiates the OAuth flow.
// STEP 1: OAuth Flow - Redirect to Google consent screen
//
// Security: Generates a random state parameter to prevent CSRF attacks.
// The state is verified when Google redirects back to our callback.
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state parameter for CSRF protection
	state := h.oauthService.GenerateState()

	// Store state in cookie for verification later
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   true, // Requires HTTPS in production
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect user to Google's OAuth consent screen
	authURL := h.oauthService.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the OAuth callback from Google.
// STEP 2-6: OAuth Flow - Handle callback, exchange code, get user info, issue JWT
//
// Flow:
// 1. Validate state parameter (CSRF protection)
// 2. Exchange authorization code for access token
// 3. Fetch user profile from Google's UserInfo endpoint
// 4. Find or create user in database
// 5. Generate JWT token
// 6. Set JWT in secure, httpOnly cookie
// 7. Redirect to frontend
func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Get state from query parameter and cookie
	state := r.URL.Query().Get("state")
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		http.Error(w, `{"error": "Invalid state parameter"}`, http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Get authorization code from query parameter
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error": "Authorization code not provided"}`, http.StatusBadRequest)
		return
	}

	// STEP 2: Exchange authorization code for access token
	token, err := h.oauthService.ExchangeCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error": "Failed to exchange code"}`, http.StatusInternalServerError)
		return
	}

	// STEP 3: Fetch user profile from Google
	userInfo, err := h.oauthService.GetUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, `{"error": "Failed to get user info"}`, http.StatusInternalServerError)
		return
	}

	// STEP 4: Find or create user in database
	user, err := h.userRepository.FindByEmail(r.Context(), userInfo.Email)
	if err != nil {
		// User doesn't exist, create new user
		user = &models.User{
			Email:   userInfo.Email,
			Name:    userInfo.Name,
			Picture: userInfo.Picture,
		}
		user, err = h.userRepository.Create(r.Context(), user)
		if err != nil {
			http.Error(w, `{"error": "Failed to create user"}`, http.StatusInternalServerError)
			return
		}
	}

	// STEP 5: Generate JWT token
	jwtToken, expiration, err := h.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, `{"error": "Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	// STEP 6: Set JWT in secure, httpOnly cookie
	// HttpOnly: Prevents JavaScript access (XSS protection)
	// Secure: Only sent over HTTPS
	// SameSite=Lax: CSRF protection for cross-site requests
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    jwtToken,
		Path:     "/",
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// STEP 7: Redirect to frontend home page
	http.Redirect(w, r, h.frontendURL+"/home", http.StatusTemporaryRedirect)
}

// Logout clears the authentication cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// UserHandler handles user-related requests.
type UserHandler struct {
	userRepository repository.UserRepository
}

// NewUserHandler creates a new user handler.
func NewUserHandler(userRepository repository.UserRepository) *UserHandler {
	return &UserHandler{userRepository: userRepository}
}

// RegisterRoutes registers user-related routes.
func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/me", h.GetMe)
}

// GetMe returns the currently logged-in user's information.
// Requires authentication via AuthMiddleware.
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by AuthMiddleware)
	userID := middleware.GetUserID(r.Context())
	if userID == 0 {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Fetch user from database
	user, err := h.userRepository.FindByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Return user info (exclude sensitive fields)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"picture":    user.Picture,
		"created_at": user.CreatedAt,
	})
}
