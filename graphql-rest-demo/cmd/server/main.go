// Package main is the entry point for the GraphQL + REST demo server.
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"graphql-rest-demo/internal/handler"
	"graphql-rest-demo/internal/repository"
	"graphql-rest-demo/internal/service"
	gql "graphql-rest-demo/pkg/graphql"
)

func main() {
	// Initialize layers
	repo := repository.NewBlogRepository()
	svc := service.NewBlogService(repo)

	// Initialize handlers
	restHandler := handler.NewRESTHandler(svc)
	gqlSchema, err := gql.NewSchema(svc)
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}

	// Create router
	mux := http.NewServeMux()

	// REST API routes
	mux.HandleFunc("/blogs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			restHandler.ListBlogs(w, r)
		case http.MethodPost:
			restHandler.CreateBlog(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	mux.HandleFunc("/blogs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			restHandler.GetBlog(w, r)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			restHandler.SearchBlogs(w, r)
		} else {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})

	// GraphQL endpoint
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "POST required"})
			return
		}

		var req struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON: " + err.Error()})
			return
		}

		result := gqlSchema.Execute(req.Query, req.Variables)
		writeJSON(w, http.StatusOK, result)
	})

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Static files
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Apply middleware
	handler := withCORS(mux)
	handler = withLogging(handler)

	// Start server
	log.Println("Server starting on http://localhost:8090")
	log.Println("REST API:")
	log.Println("  - GET  /blogs?author=&tag=&q=")
	log.Println("  - GET  /blogs/{id}")
	log.Println("  - POST /blogs")
	log.Println("  - GET  /search?q=")
	log.Println("GraphQL API:")
	log.Println("  - POST /graphql")
	log.Println("Web UI: http://localhost:8090")
	log.Fatal(http.ListenAndServe(":8090", handler))
}

// Middleware

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// Handle GraphQL result specially
	if result, ok := data.(*struct {
		Data   interface{}              `json:"data,omitempty"`
		Errors []map[string]interface{} `json:"errors,omitempty"`
	}); ok {
		json.NewEncoder(w).Encode(result)
		return
	}

	json.NewEncoder(w).Encode(data)
}
