// Package main is the entry point for the Elasticsearch demo server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"elasticsearch-demo/internal/handler"
	"elasticsearch-demo/internal/repository"
	"elasticsearch-demo/internal/service"
	"elasticsearch-demo/pkg/elasticsearch"
)

func main() {
	// Initialize Elasticsearch client
	log.Println("Connecting to Elasticsearch...")
	esClient, err := elasticsearch.NewClient("")
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	log.Println("Connected to Elasticsearch successfully")

	// Initialize layers
	blogRepo := repository.NewBlogRepository(esClient)
	blogService := service.NewBlogService(blogRepo)
	blogHandler := handler.NewBlogHandler(blogService)

	// Create router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/search", blogHandler.Search)
	mux.HandleFunc("/autocomplete", blogHandler.Autocomplete)
	mux.HandleFunc("/blogs", blogHandler.CreateBlog)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static files (frontend)
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Apply CORS middleware
	handler := withCORS(mux)

	// Create server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8050"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on http://localhost:%s", port)
		log.Printf("API endpoints:")
		log.Printf("  - GET  /search?q=...\u0026author=...\u0026tag=...\u0026page=1\u0026size=10\u0026sort=newest")
		log.Printf("  - GET  /autocomplete?q=...")
		log.Printf("  - POST /blogs")
		log.Printf("  - GET  /health")
		log.Printf("Frontend: http://localhost:%s/", port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}

// withCORS wraps the handler with CORS headers for cross-origin requests.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
