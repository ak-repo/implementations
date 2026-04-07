// Package model defines the domain entities for the blog system.
package model

import "time"

// Blog represents a blog post.
type Blog struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBlogRequest represents the request to create a blog.
type CreateBlogRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Author  string   `json:"author"`
	Tags    []string `json:"tags"`
}

// BlogFilter represents filter options for listing blogs.
type BlogFilter struct {
	Author string
	Tag    string
	Query  string
}
