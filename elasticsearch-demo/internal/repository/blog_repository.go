// Package repository provides data access operations for blogs using Elasticsearch.
package repository

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"elasticsearch-demo/internal/model"
	"elasticsearch-demo/pkg/elasticsearch"
)

// BlogRepository handles Elasticsearch operations for blogs.
type BlogRepository struct {
	esClient *elasticsearch.Client
}

// NewBlogRepository creates a new blog repository.
func NewBlogRepository(esClient *elasticsearch.Client) *BlogRepository {
	return &BlogRepository{
		esClient: esClient,
	}
}

// IndexBlog indexes a single blog post into Elasticsearch.
func (r *BlogRepository) IndexBlog(blog *model.Blog) error {
	if blog.ID == "" {
		blog.ID = generateID()
	}
	if blog.CreatedAt.IsZero() {
		blog.CreatedAt = time.Now()
	}

	return r.esClient.Index("blogs", blog.ID, blog)
}

// SearchBlogs performs a full-text search with filters, pagination, sorting, and highlighting.
func (r *BlogRepository) SearchBlogs(req model.SearchRequest) (*model.HighlightedSearchResult, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 || req.Size > 100 {
		req.Size = 10
	}

	// Build the bool query
	boolQuery := map[string]interface{}{
		"must":   []map[string]interface{}{},
		"filter": []map[string]interface{}{},
	}

	// Add full-text search if query provided
	if req.Query != "" {
		boolQuery["must"] = append(boolQuery["must"].([]map[string]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  req.Query,
				"fields": []string{"title^3", "content"}, // Boost title by 3x
				"type":   "best_fields",
			},
		})
	} else {
		// Match all if no query
		boolQuery["must"] = append(boolQuery["must"].([]map[string]interface{}), map[string]interface{}{
			"match_all": map[string]interface{}{},
		})
	}

	// Add author filter
	if req.Author != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"author": req.Author,
			},
		})
	}

	// Add tag filter
	if req.Tag != "" {
		boolQuery["filter"] = append(boolQuery["filter"].([]map[string]interface{}), map[string]interface{}{
			"term": map[string]interface{}{
				"tags": req.Tag,
			},
		})
	}

	// Build sort
	sort := buildSort(req.SortBy)

	// Calculate from for pagination
	from := (req.Page - 1) * req.Size

	// Build the complete search query
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"sort":      sort,
		"from":      from,
		"size":      req.Size,
		"highlight": buildHighlight(),
		"_source":   true,
	}

	// Execute search
	response, err := r.esClient.Search("blogs", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return r.parseSearchResponse(response, req.Page, req.Size)
}

// Autocomplete provides search suggestions based on title prefix matching.
func (r *BlogRepository) Autocomplete(query string) ([]string, error) {
	if query == "" {
		return []string{}, nil
	}

	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase_prefix": map[string]interface{}{
				"title": map[string]interface{}{
					"query":          query,
					"max_expansions": 10,
				},
			},
		},
		"size":    5,
		"_source": []string{"title"},
	}

	response, err := r.esClient.Search("blogs", searchQuery)
	if err != nil {
		return nil, fmt.Errorf("autocomplete failed: %w", err)
	}

	return r.parseAutocompleteResponse(response)
}

// buildSort creates the sort configuration based on sort parameter.
func buildSort(sortBy string) []map[string]interface{} {
	switch sortBy {
	case "newest":
		return []map[string]interface{}{
			{"created_at": map[string]string{"order": "desc"}},
			{"_score": map[string]string{"order": "desc"}},
		}
	case "oldest":
		return []map[string]interface{}{
			{"created_at": map[string]string{"order": "asc"}},
			{"_score": map[string]string{"order": "desc"}},
		}
	default:
		// relevance - sort by score first
		return []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"created_at": map[string]string{"order": "desc"}},
		}
	}
}

// buildHighlight creates the highlight configuration.
func buildHighlight() map[string]interface{} {
	return map[string]interface{}{
		"fields": map[string]interface{}{
			"title": map[string]interface{}{
				"fragment_size":       100,
				"number_of_fragments": 1,
			},
			"content": map[string]interface{}{
				"fragment_size":       200,
				"number_of_fragments": 2,
			},
		},
		"pre_tags":  []string{"<em class=\"highlight\">"},
		"post_tags": []string{"</em>"},
	}
}

// parseSearchResponse converts Elasticsearch response to our model.
func (r *BlogRepository) parseSearchResponse(response map[string]interface{}, page, size int) (*model.HighlightedSearchResult, error) {
	hits := response["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	hitsArray := hits["hits"].([]interface{})

	blogs := make([]*model.HighlightedBlog, 0, len(hitsArray))

	for _, hit := range hitsArray {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		blog := &model.HighlightedBlog{
			Blog: model.Blog{
				ID:      hitMap["_id"].(string),
				Title:   getString(source, "title"),
				Content: getString(source, "content"),
				Author:  getString(source, "author"),
				Tags:    getStringSlice(source, "tags"),
			},
			Highlights: make(map[string][]string),
		}

		// Parse created_at
		if createdAtStr, ok := source["created_at"].(string); ok {
			if t, err := time.Parse(time.RFC3339, createdAtStr); err == nil {
				blog.Blog.CreatedAt = t
			}
		}

		// Parse highlights
		if highlight, ok := hitMap["highlight"].(map[string]interface{}); ok {
			for field, fragments := range highlight {
				if fragmentsSlice, ok := fragments.([]interface{}); ok {
					fragmentsStr := make([]string, len(fragmentsSlice))
					for i, f := range fragmentsSlice {
						fragmentsStr[i] = f.(string)
					}
					blog.Highlights[field] = fragmentsStr
				}
			}
		}

		blogs = append(blogs, blog)
	}

	return &model.HighlightedSearchResult{
		Total: int64(total),
		Page:  page,
		Size:  size,
		Blogs: blogs,
	}, nil
}

// parseAutocompleteResponse extracts suggestions from Elasticsearch response.
func (r *BlogRepository) parseAutocompleteResponse(response map[string]interface{}) ([]string, error) {
	hits := response["hits"].(map[string]interface{})
	hitsArray := hits["hits"].([]interface{})

	suggestions := make([]string, 0, len(hitsArray))
	for _, hit := range hitsArray {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		if title, ok := source["title"].(string); ok {
			suggestions = append(suggestions, title)
		}
	}

	return suggestions, nil
}

// Helper functions
func generateID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if v, ok := m[key].([]interface{}); ok {
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	// Try to parse from string if it's a comma-separated list
	if v, ok := m[key].(string); ok {
		return strings.Split(v, ",")
	}
	return []string{}
}
