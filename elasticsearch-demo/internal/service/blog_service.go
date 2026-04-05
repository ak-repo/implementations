// Package service provides business logic for the blog search system.
package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"elasticsearch-demo/internal/model"
	"elasticsearch-demo/internal/repository"
)

// BlogService handles business logic for blog operations.
type BlogService struct {
	repo *repository.BlogRepository
}

// NewBlogService creates a new blog service.
func NewBlogService(repo *repository.BlogRepository) *BlogService {
	return &BlogService{
		repo: repo,
	}
}

// CreateBlogRequest represents the request body for creating a blog.
type CreateBlogRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Author  string   `json:"author"`
	Tags    []string `json:"tags"`
}

// SearchBlogs performs a search with validation and business logic.
func (s *BlogService) SearchBlogs(query, author, tag, page, size, sortBy string) (*model.HighlightedSearchResult, error) {
	// Parse and validate pagination
	pageNum := parseInt(page, 1)
	sizeNum := parseInt(size, 10)
	if sizeNum > 100 {
		sizeNum = 100 // Max page size
	}

	// Validate sort parameter
	validSorts := map[string]bool{
		"relevance": true,
		"newest":    true,
		"oldest":    true,
	}
	if !validSorts[sortBy] {
		sortBy = "relevance"
	}

	// Sanitize inputs
	searchReq := model.SearchRequest{
		Query:  strings.TrimSpace(query),
		Author: strings.TrimSpace(author),
		Tag:    strings.TrimSpace(tag),
		Page:   pageNum,
		Size:   sizeNum,
		SortBy: sortBy,
	}

	return s.repo.SearchBlogs(searchReq)
}

// Autocomplete provides search suggestions.
func (s *BlogService) Autocomplete(query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return []string{}, nil // Require at least 2 characters
	}
	return s.repo.Autocomplete(query)
}

// CreateBlog creates a new blog post after validation.
func (s *BlogService) CreateBlog(req *CreateBlogRequest) (*model.Blog, error) {
	// Validation
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Sanitize
	blog := &model.Blog{
		Title:   strings.TrimSpace(req.Title),
		Content: strings.TrimSpace(req.Content),
		Author:  strings.TrimSpace(req.Author),
		Tags:    sanitizeTags(req.Tags),
	}

	if err := s.repo.IndexBlog(blog); err != nil {
		return nil, fmt.Errorf("failed to index blog: %w", err)
	}

	return blog, nil
}

// validateCreateRequest validates the create blog request.
func validateCreateRequest(req *CreateBlogRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return &ValidationError{Field: "title", Message: "title is required"}
	}
	if len(req.Title) > 200 {
		return &ValidationError{Field: "title", Message: "title must be less than 200 characters"}
	}
	if strings.TrimSpace(req.Content) == "" {
		return &ValidationError{Field: "content", Message: "content is required"}
	}
	if strings.TrimSpace(req.Author) == "" {
		return &ValidationError{Field: "author", Message: "author is required"}
	}
	if len(req.Author) > 100 {
		return &ValidationError{Field: "author", Message: "author must be less than 100 characters"}
	}
	return nil
}

// sanitizeTags removes empty tags and deduplicates.
func sanitizeTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag != "" && !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}
	return result
}

// parseInt parses an integer from string with a default value.
func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil || result < 1 {
		return defaultVal
	}
	return result
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// WriteError writes an error response in JSON format.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{
		"error": message,
	})
}
