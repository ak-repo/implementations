// Package handler provides HTTP handlers for the blog search API.
package handler

import (
	"encoding/json"
	"net/http"

	"elasticsearch-demo/internal/service"
)

// BlogHandler handles HTTP requests for blog operations.
type BlogHandler struct {
	service *service.BlogService
}

// NewBlogHandler creates a new blog handler.
func NewBlogHandler(service *service.BlogService) *BlogHandler {
	return &BlogHandler{
		service: service,
	}
}

// SearchResponse represents the response structure for search endpoint.
type SearchResponse struct {
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	Size  int        `json:"size"`
	Blogs []BlogItem `json:"blogs"`
}

// BlogItem represents a single blog in the response.
type BlogItem struct {
	ID         string              `json:"id"`
	Title      string              `json:"title"`
	Content    string              `json:"content"`
	Author     string              `json:"author"`
	Tags       []string            `json:"tags"`
	CreatedAt  string              `json:"created_at"`
	Highlights map[string][]string `json:"highlights,omitempty"`
}

// Search handles GET /search - searches blogs with filters, pagination, and highlighting.
func (h *BlogHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		service.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Extract query parameters
	query := r.URL.Query().Get("q")
	author := r.URL.Query().Get("author")
	tag := r.URL.Query().Get("tag")
	page := r.URL.Query().Get("page")
	size := r.URL.Query().Get("size")
	sort := r.URL.Query().Get("sort")

	// Perform search
	result, err := h.service.SearchBlogs(query, author, tag, page, size, sort)
	if err != nil {
		service.WriteError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	// Transform to response format
	blogs := make([]BlogItem, len(result.Blogs))
	for i, b := range result.Blogs {
		blogs[i] = BlogItem{
			ID:         b.ID,
			Title:      b.Title,
			Content:    b.Content,
			Author:     b.Author,
			Tags:       b.Tags,
			CreatedAt:  b.CreatedAt.Format("2006-01-02T15:04:05Z"),
			Highlights: b.Highlights,
		}
		// Use highlighted title/content if available
		if len(b.Highlights["title"]) > 0 {
			blogs[i].Title = b.Highlights["title"][0]
		}
		if len(b.Highlights["content"]) > 0 {
			blogs[i].Content = joinHighlights(b.Highlights["content"])
		}
	}

	response := SearchResponse{
		Total: result.Total,
		Page:  result.Page,
		Size:  result.Size,
		Blogs: blogs,
	}

	service.WriteJSON(w, http.StatusOK, response)
}

// Autocomplete handles GET /autocomplete - provides search suggestions.
func (h *BlogHandler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		service.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		service.WriteJSON(w, http.StatusOK, map[string][]string{"suggestions": {}})
		return
	}

	suggestions, err := h.service.Autocomplete(query)
	if err != nil {
		service.WriteError(w, http.StatusInternalServerError, "autocomplete failed: "+err.Error())
		return
	}

	service.WriteJSON(w, http.StatusOK, map[string][]string{"suggestions": suggestions})
}

// CreateBlog handles POST /blogs - creates a new blog post.
func (h *BlogHandler) CreateBlog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		service.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.CreateBlogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		service.WriteError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	blog, err := h.service.CreateBlog(&req)
	if err != nil {
		if validationErr, ok := err.(*service.ValidationError); ok {
			service.WriteJSON(w, http.StatusBadRequest, map[string]interface{}{
				"error":   "validation failed",
				"field":   validationErr.Field,
				"message": validationErr.Message,
			})
			return
		}
		service.WriteError(w, http.StatusInternalServerError, "failed to create blog: "+err.Error())
		return
	}

	service.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         blog.ID,
		"title":      blog.Title,
		"author":     blog.Author,
		"tags":       blog.Tags,
		"created_at": blog.CreatedAt.Format("2006-01-02T15:04:05Z"),
	})
}

// joinHighlights joins multiple highlight fragments.
func joinHighlights(highlights []string) string {
	if len(highlights) == 0 {
		return ""
	}
	if len(highlights) == 1 {
		return highlights[0]
	}
	result := highlights[0]
	for i := 1; i < len(highlights); i++ {
		result += " ... " + highlights[i]
	}
	return result
}
