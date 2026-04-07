// Package handler provides REST HTTP handlers.
package handler

import (
	"encoding/json"
	"graphql-rest-demo/internal/model"
	"graphql-rest-demo/internal/service"
	"net/http"
	"strings"
)

// RESTHandler handles REST API requests.
type RESTHandler struct {
	service *service.BlogService
}

// NewRESTHandler creates a new REST handler.
func NewRESTHandler(service *service.BlogService) *RESTHandler {
	return &RESTHandler{service: service}
}

// ListBlogs handles GET /blogs with optional filters.
func (h *RESTHandler) ListBlogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	filter := model.BlogFilter{
		Author: r.URL.Query().Get("author"),
		Tag:    r.URL.Query().Get("tag"),
		Query:  r.URL.Query().Get("q"),
	}

	blogs := h.service.ListBlogs(filter)
	writeJSON(w, http.StatusOK, blogs)
}

// GetBlog handles GET /blogs/{id}.
func (h *RESTHandler) GetBlog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract ID from path: /blogs/{id}
	path := strings.TrimPrefix(r.URL.Path, "/blogs/")
	id := strings.Split(path, "/")[0]

	if id == "" {
		writeError(w, http.StatusBadRequest, "blog ID required")
		return
	}

	blog, err := h.service.GetBlog(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, blog)
}

// CreateBlog handles POST /blogs.
func (h *RESTHandler) CreateBlog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.CreateBlogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	blog, err := h.service.CreateBlog(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, blog)
}

// SearchBlogs handles GET /search.
func (h *RESTHandler) SearchBlogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' required")
		return
	}

	blogs := h.service.SearchBlogs(query)
	writeJSON(w, http.StatusOK, blogs)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
