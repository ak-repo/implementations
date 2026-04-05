// Package elasticsearch provides the Elasticsearch client initialization and configuration.
package elasticsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// Client wraps the Elasticsearch HTTP client with retry logic.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Elasticsearch client with the provided URL.
// If url is empty, it reads from ELASTICSEARCH_URL environment variable.
func NewClient(url string) (*Client, error) {
	if url == "" {
		url = os.Getenv("ELASTICSEARCH_URL")
		if url == "" {
			url = "http://localhost:9200"
		}
	}

	client := &Client{
		baseURL: url,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Wait for Elasticsearch to be ready with retries
	if err := client.waitForReady(5); err != nil {
		return nil, fmt.Errorf("elasticsearch not ready: %w", err)
	}

	// Initialize index and mappings
	if err := client.InitializeIndex(); err != nil {
		return nil, fmt.Errorf("failed to initialize index: %w", err)
	}

	return client, nil
}

// waitForReady polls Elasticsearch until it's ready or max retries reached.
func (c *Client) waitForReady(maxRetries int) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := c.httpClient.Get(c.baseURL + "/_cluster/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("elasticsearch not available after %d retries", maxRetries)
}

// InitializeIndex creates the blogs index with proper mappings if it doesn't exist.
func (c *Client) InitializeIndex() error {
	indexURL := fmt.Sprintf("%s/blogs", c.baseURL)

	// Check if index exists
	resp, err := c.httpClient.Head(indexURL)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	resp.Body.Close()

	// Index already exists
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Create index with mappings
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]string{
					"type": "text",
				},
				"content": map[string]string{
					"type": "text",
				},
				"author": map[string]string{
					"type": "keyword",
				},
				"tags": map[string]string{
					"type": "keyword",
				},
				"created_at": map[string]string{
					"type": "date",
				},
			},
		},
	}

	jsonBody, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, indexURL, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create index, status: %d, body: %s", resp.StatusCode, string(body))
	}

	fmt.Println("Created 'blogs' index with mappings")
	return nil
}

// Index creates or updates a document in the specified index.
func (c *Client) Index(index string, id string, doc interface{}) error {
	url := fmt.Sprintf("%s/%s/_doc/%s", c.baseURL, index, id)

	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to index document, status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Search executes a search query against the specified index.
func (c *Client) Search(index string, query map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s/_search", c.baseURL, index)

	jsonBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

// HTTPClient returns the underlying HTTP client for direct access if needed.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// BaseURL returns the Elasticsearch base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}
