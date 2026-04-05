// cmd/server/main.go
// Main entry point for the OAuth 2.0 demo backend server.
//
// This server implements:
// - Google OAuth 2.0 authentication flow
// - JWT-based session management with secure cookies
// - SQLite database for user storage
// - RESTful API endpoints for authentication
//
// Environment variables required:
// - GOOGLE_CLIENT_ID: Google OAuth client ID
// - GOOGLE_CLIENT_SECRET: Google OAuth client secret
// - GOOGLE_REDIRECT_URL: Callback URL for OAuth (e.g., http://localhost:8090/auth/google/callback)
// - JWT_SECRET: Secret key for JWT signing (minimum 32 characters)
// - FRONTEND_URL: URL of the frontend application (e.g., http://localhost:3050)
// - PORT: Server port (default: 8090)
// - DB_PATH: Path to SQLite database (default: ./users.db)

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"oauth-demo/internal/handlers"
	oauthMiddleware "oauth-demo/internal/middleware"
	"oauth-demo/internal/repository"
	svc "oauth-demo/internal/services"
	"oauth-demo/pkg/database"
)

func main() {
	// Load environment variables from .env file
	// Ignore error if .env doesn't exist (for production deployments)
	_ = godotenv.Load()

	// Initialize database
	dbPath := getEnv("DB_PATH", "./users.db")
	db, err := database.NewSQLiteDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewSQLiteUserRepository(db)

	// Initialize services
	jwtService, err := svc.NewJWTService()
	if err != nil {
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}

	oauthService, err := svc.NewOAuthService()
	if err != nil {
		log.Fatalf("Failed to initialize OAuth service: %v", err)
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(oauthService, jwtService, userRepo)
	userHandler := handlers.NewUserHandler(userRepo)

	// Create router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)    // Request logging
	r.Use(middleware.Recoverer) // Panic recovery
	r.Use(middleware.RequestID) // Add request ID to each request

	// CORS configuration
	// Allow credentials (cookies) and specific origins
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{getEnv("FRONTEND_URL", "http://localhost:3050")},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Required for cookies
		MaxAge:           300,
	}))

	// Public routes (no auth required)
	r.Group(func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(oauthMiddleware.AuthMiddleware(jwtService))
		userHandler.RegisterRoutes(r)
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// Start server
	port := getEnv("PORT", "8090")
	log.Printf("Server starting on port %s", port)
	log.Printf("Frontend URL: %s", getEnv("FRONTEND_URL", "http://localhost:3050"))
	log.Fatal(http.ListenAndServe(":"+port, r))
}

// getEnv retrieves an environment variable with a default fallback.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
