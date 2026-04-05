# Elasticsearch Blog Search Demo

A production-style full-text search system built with **Go** and **Elasticsearch 8.x**, featuring autocomplete, highlighting, filtering, and pagination.

## Features

- **Full-text Search**: Multi-field search with relevance scoring (title boosted 3x)
- **Filters**: Filter by author and tags
- **Sorting**: Sort by relevance, newest, or oldest
- **Highlighting**: Search terms highlighted in results
- **Autocomplete**: Real-time search suggestions
- **Pagination**: Efficient offset-based pagination
- **Clean Architecture**: Handler → Service → Repository pattern

## Architecture

```
cmd/server/
└── main.go              # Entry point, HTTP server

internal/
├── handler/
│   └── blog_handler.go  # HTTP handlers (Search, Autocomplete, Create)
├── service/
│   └── blog_service.go  # Business logic & validation
├── repository/
│   └── blog_repository.go  # Elasticsearch queries
└── model/
    └── blog.go          # Domain models

pkg/elasticsearch/
└── client.go            # ES client initialization & index setup

web/
├── index.html           # Frontend UI
└── app.js               # Search & autocomplete logic

scripts/
└── seed.go              # Sample data generator
```

## Why Elasticsearch?

| Feature | SQL LIKE | Elasticsearch |
|---------|----------|---------------|
| Relevance Scoring | ❌ No | ✅ TF-IDF based |
| Multi-field Search | ❌ Complex | ✅ Native multi_match |
| Typo Tolerance | ❌ No | ✅ Fuzzy matching ready |
| Highlighting | ❌ Manual | ✅ Built-in |
| Full-text Analytics | ❌ No | ✅ Aggregations |
| Performance | ⚠️ Table scan | ✅ Inverted index |

## Quick Start

### 1. Start Elasticsearch

```bash
docker-compose up -d
```

Wait for Elasticsearch to be ready (~30 seconds):
```bash
curl http://localhost:9200/_cluster/health
```

### 2. Run the Server

```bash
go run cmd/server/main.go
```

The server will:
- Connect to Elasticsearch
- Create the `blogs` index with mappings
- Start on `http://localhost:8080`

### 3. Seed Sample Data

In a new terminal:
```bash
go run scripts/seed.go
```

This creates 100 sample blog posts with realistic titles, authors, and tags.

### 4. Open the UI

Visit: `http://localhost:8080`

## API Reference

### Search Blogs
```
GET /search?q={query}&author={author}&tag={tag}&page={page}&size={size}&sort={sort}
```

**Parameters:**
- `q` - Search query (searches title + content)
- `author` - Filter by author name
- `tag` - Filter by tag
- `page` - Page number (default: 1)
- `size` - Results per page, max 100 (default: 10)
- `sort` - `relevance`, `newest`, or `oldest` (default: relevance)

**Example:**
```bash
curl "http://localhost:8080/search?q=golang&author=Alex%20Johnson&tag=backend&sort=newest"
```

**Response:**
```json
{
  "total": 42,
  "page": 1,
  "size": 10,
  "blogs": [
    {
      "id": "12345",
      "title": "Getting Started with <em class=\"highlight\">Golang</em>",
      "content": "In this guide we explore <em class=\"highlight\">Golang</em>...",
      "author": "Alex Johnson",
      "tags": ["golang", "backend"],
      "created_at": "2024-01-15T10:30:00Z",
      "highlights": {
        "title": ["Getting Started with <em class=\"highlight\">Golang</em>"],
        "content": ["In this guide we explore <em class=\"highlight\">Golang</em>..."]
      }
    }
  ]
}
```

### Autocomplete
```
GET /autocomplete?q={query}
```

**Example:**
```bash
curl "http://localhost:8080/autocomplete?q=go"
```

**Response:**
```json
{
  "suggestions": ["golang tutorial", "golang concurrency", "go patterns"]
}
```

### Create Blog
```
POST /blogs
Content-Type: application/json
```

**Body:**
```json
{
  "title": "My Blog Post",
  "content": "This is the full content...",
  "author": "John Doe",
  "tags": ["golang", "tutorial"]
}
```

## Elasticsearch Query Structure

The search uses a `bool` query combining:

```json
{
  "query": {
    "bool": {
      "must": [
        {
          "multi_match": {
            "query": "search terms",
            "fields": ["title^3", "content"],
            "type": "best_fields"
          }
        }
      ],
      "filter": [
        {"term": {"author": "John"}},
        {"term": {"tags": "golang"}}
      ]
    }
  },
  "highlight": {
    "fields": {
      "title": {"fragment_size": 100},
      "content": {"fragment_size": 200}
    }
  }
}
```

**Key Points:**
- `must` - Affects relevance scoring
- `filter` - Cached, no scoring (better performance)
- Title is boosted 3x over content
- Highlights are HTML-escaped automatically

## Index Mapping

```json
{
  "mappings": {
    "properties": {
      "title": {"type": "text"},
      "content": {"type": "text"},
      "author": {"type": "keyword"},
      "tags": {"type": "keyword"},
      "created_at": {"type": "date"}
    }
  }
}
```

**Field Types Explained:**
- `text` - Full-text searchable, analyzed (tokenized, lowercased)
- `keyword` - Exact match, not analyzed (good for filtering)
- `date` - Date range queries and sorting

## Configuration

Environment variables (see `.env.example`):

```bash
ELASTICSEARCH_URL=http://localhost:9200  # Elasticsearch endpoint
PORT=8080                                 # Server port
```

## Stopping the Demo

```bash
# Stop the server (Ctrl+C in the terminal)

# Stop Elasticsearch
docker-compose down

# Remove Elasticsearch data (optional)
docker-compose down -v
```

## Troubleshooting

### Connection refused to Elasticsearch
```bash
# Check if Elasticsearch is running
docker ps | grep elasticsearch

# Check logs
docker-compose logs elasticsearch

# Wait for it to be ready
curl http://localhost:9200/_cluster/health
```

### No search results
```bash
# Check if data exists
curl http://localhost:9200/blogs/_count

# Re-seed the data
go run scripts/seed.go
```

### CORS errors in browser
The server includes CORS headers. If issues persist, check that you're accessing `http://localhost:8080` (not `127.0.0.1` or other variants).

## Extending the Demo

### Add Faceted Search
Index aggregations to get filter counts:
```json
{
  "aggs": {
    "authors": {"terms": {"field": "author"}},
    "tags": {"terms": {"field": "tags"}}
  }
}
```

### Add Fuzzy Matching
For typo-tolerant search:
```json
{
  "multi_match": {
    "query": "golan",
    "fields": ["title", "content"],
    "fuzziness": "AUTO"
  }
}
```

### Implement Suggestions
Use the Elasticsearch `_suggest` endpoint for better autocomplete with completion suggester.

## Resources

- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Go Elasticsearch Client](https://github.com/elastic/go-elasticsearch)
- [Query DSL Reference](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl.html)

## License

MIT - Feel free to use for your projects!