// Package middleware provides HTTP middleware for authentication and authorization.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"oauth-demo/internal/services"
)

// Context keys for storing values in request context.
type contextKey string

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyEmail    contextKey = "email"
	ContextKeyUserName contextKey = "user_name"
)

// AuthMiddleware validates JWT tokens from cookies or Authorization header.
// It extracts the user info and adds it to the request context.
func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from cookie or Authorization header
			var tokenString string

			// First, try to get from cookie
			cookie, err := r.Cookie("auth_token")
			if err == nil {
				tokenString = cookie.Value
			} else {
				// Fallback to Authorization header
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenString == "" {
				http.Error(w, `{"error": "Unauthorized - no token provided"}`, http.StatusUnauthorized)
				return
			}

			// Validate the token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, `{"error": "Unauthorized - invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyEmail, claims.Email)
			ctx = context.WithValue(ctx, ContextKeyUserName, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts user ID from request context.
// Returns 0 if not found.
func GetUserID(ctx context.Context) int64 {
	if id, ok := ctx.Value(ContextKeyUserID).(int64); ok {
		return id
	}
	return 0
}

// GetUserEmail extracts user email from request context.
func GetUserEmail(ctx context.Context) string {
	if email, ok := ctx.Value(ContextKeyEmail).(string); ok {
		return email
	}
	return ""
}

// GetUserName extracts user name from request context.
func GetUserName(ctx context.Context) string {
	if name, ok := ctx.Value(ContextKeyUserName).(string); ok {
		return name
	}
	return ""
}
