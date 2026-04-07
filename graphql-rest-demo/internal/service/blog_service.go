// Package service provides business logic for the blog system.
package service

import (
	"errors"
	"graphql-rest-demo/internal/model"
	"graphql-rest-demo/internal/repository"
	"strings"
)

// BlogService handles business logic for blogs.
type BlogService struct {
	repo *repository.BlogRepository
}

// NewBlogService creates a new blog service.
func NewBlogService(repo *repository.BlogRepository) *BlogService {
	return &BlogService{repo: repo}
}

// CreateBlog creates a new blog with validation.
func (s *BlogService) CreateBlog(req model.CreateBlogRequest) (*model.Blog, error) {
	// Validate
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

	if err := s.repo.Create(blog); err != nil {
		return nil, err
	}

	return blog, nil
}

// GetBlog retrieves a blog by ID.
func (s *BlogService) GetBlog(id string) (*model.Blog, error) {
	return s.repo.GetByID(id)
}

// ListBlogs returns all blogs with optional filtering.
func (s *BlogService) ListBlogs(filter model.BlogFilter) []*model.Blog {
	return s.repo.Search(filter)
}

// SearchBlogs searches blogs by query string.
func (s *BlogService) SearchBlogs(query string) []*model.Blog {
	return s.repo.Search(model.BlogFilter{Query: query})
}

func validateCreateRequest(req model.CreateBlogRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return errors.New("title is required")
	}
	if len(req.Title) > 200 {
		return errors.New("title must be less than 200 characters")
	}
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("content is required")
	}
	if strings.TrimSpace(req.Author) == "" {
		return errors.New("author is required")
	}
	return nil
}

func sanitizeTags(tags []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(tags))

	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag != "" && !seen[tag] {
			seen[tag] = true
			result = append(result, tag)
		}
	}

	return result
}
