// Package model defines the domain entities for the blog search system.
package model

import "time"

// Blog represents a blog post in the search system.
type Blog struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchRequest contains all parameters for searching blogs.
type SearchRequest struct {
	Query  string
	Author string
	Tag    string
	Page   int
	Size   int
	SortBy string // "relevance" or "newest"
}

// SearchResult contains the paginated search results.
type SearchResult struct {
	Total int64   `json:"total"`
	Page  int     `json:"page"`
	Size  int     `json:"size"`
	Blogs []*Blog `json:"blogs"`
}

// HighlightedBlog extends Blog with highlighted fields.
type HighlightedBlog struct {
	Blog
	Highlights map[string][]string `json:"highlights,omitempty"`
}

// HighlightedSearchResult contains search results with highlighting.
type HighlightedSearchResult struct {
	Total int64              `json:"total"`
	Page  int                `json:"page"`
	Size  int                `json:"size"`
	Blogs []*HighlightedBlog `json:"blogs"`
}
