// Package repository provides in-memory storage for blogs.
package repository

import (
	"errors"
	"graphql-rest-demo/internal/model"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BlogRepository stores blogs in memory.
type BlogRepository struct {
	blogs map[string]*model.Blog
	mu    sync.RWMutex
}

// NewBlogRepository creates a new repository with sample data.
func NewBlogRepository() *BlogRepository {
	r := &BlogRepository{
		blogs: make(map[string]*model.Blog),
	}

	// Seed sample data
	r.seedData()
	return r
}

// Create adds a new blog to the repository.
func (r *BlogRepository) Create(blog *model.Blog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	blog.ID = uuid.New().String()
	blog.CreatedAt = time.Now()

	r.blogs[blog.ID] = blog
	return nil
}

// GetByID retrieves a blog by its ID.
func (r *BlogRepository) GetByID(id string) (*model.Blog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	blog, exists := r.blogs[id]
	if !exists {
		return nil, errors.New("blog not found")
	}

	return blog, nil
}

// List returns all blogs.
func (r *BlogRepository) List() []*model.Blog {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.Blog, 0, len(r.blogs))
	for _, blog := range r.blogs {
		result = append(result, blog)
	}

	return result
}

// Search filters blogs based on criteria.
func (r *BlogRepository) Search(filter model.BlogFilter) []*model.Blog {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*model.Blog

	for _, blog := range r.blogs {
		// Apply author filter
		if filter.Author != "" && blog.Author != filter.Author {
			continue
		}

		// Apply tag filter
		if filter.Tag != "" && !contains(blog.Tags, filter.Tag) {
			continue
		}

		// Apply query filter (search in title and content)
		if filter.Query != "" {
			if !containsIgnoreCase(blog.Title, filter.Query) &&
				!containsIgnoreCase(blog.Content, filter.Query) {
				continue
			}
		}

		result = append(result, blog)
	}

	return result
}

// seedData populates initial sample data.
func (r *BlogRepository) seedData() {
	samples := []model.CreateBlogRequest{
		{
			Title:   "Getting Started with GraphQL",
			Content: "GraphQL is a query language for APIs that provides a complete and understandable description of the data...",
			Author:  "Alice Johnson",
			Tags:    []string{"graphql", "api", "tutorial"},
		},
		{
			Title:   "REST API Best Practices",
			Content: "REST has been the standard for API development for many years. Here are the best practices...",
			Author:  "Bob Smith",
			Tags:    []string{"rest", "api", "best-practices"},
		},
		{
			Title:   "Comparing GraphQL vs REST",
			Content: "When should you use GraphQL over REST? This article explores the trade-offs...",
			Author:  "Alice Johnson",
			Tags:    []string{"graphql", "rest", "comparison"},
		},
		{
			Title:   "Building Scalable APIs",
			Content: "Scalability is crucial for modern applications. Learn how to design APIs that scale...",
			Author:  "Carol White",
			Tags:    []string{"scalability", "api", "architecture"},
		},
		{
			Title:   "Authentication in GraphQL",
			Content: "Securing your GraphQL API is essential. This guide covers authentication strategies...",
			Author:  "Bob Smith",
			Tags:    []string{"graphql", "security", "authentication"},
		},
	}

	for _, sample := range samples {
		blog := &model.Blog{
			Title:     sample.Title,
			Content:   sample.Content,
			Author:    sample.Author,
			Tags:      sample.Tags,
			CreatedAt: time.Now().Add(-time.Duration(len(samples)) * time.Hour),
		}
		blog.ID = uuid.New().String()
		r.blogs[blog.ID] = blog
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return len(substr) > 0 &&
		(len(s) >= len(substr) &&
			(s == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
