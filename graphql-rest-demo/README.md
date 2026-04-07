# GraphQL vs REST Demo

A production-style demo showing the same blog data exposed via both **REST** and **GraphQL** APIs.

## What's the Difference?

### REST
- **Multiple endpoints**: `/blogs`, `/blogs/{id}`, `/search`
- **Fixed responses**: You get all fields, even if you only need a few
- **Over-fetching**: Wastes bandwidth on unused data

### GraphQL
- **Single endpoint**: `/graphql`
- **Flexible queries**: Request only the fields you need
- **Precise responses**: No wasted bandwidth

## Quick Start

### 1. Install Dependencies

```bash
go mod tidy
```

### 2. Run the Server

```bash
go run cmd/server/main.go
```

Server starts on `http://localhost:8080`

### 3. Open Web UI

```bash
open http://localhost:8080
```

## Architecture

```
┌─────────────────┐
│   Web Client    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Go Server     │
│   Port 8080     │
│                 │
│  REST Handlers ─┼── GET  /blogs
│                 │    GET  /blogs/{id}
│                 │    POST /blogs
│                 │    GET  /search
│                 │
│  GraphQL Schema─┼── POST /graphql
│                 │    Query / Mutation
│                 │
│  Service Layer──┐
│                 │
│  Repository ────┼── In-Memory Store
└─────────────────┘
```

## REST API

### List All Blogs
```bash
GET /blogs

# With filters
GET /blogs?author=Alice%20Johnson
GET /blogs?tag=graphql
GET /blogs?q=search-term
```

**Response:**
```json
[
  {
    "id": "abc123",
    "title": "Getting Started with GraphQL",
    "content": "Full content here...",
    "author": "Alice Johnson",
    "tags": ["graphql", "api", "tutorial"],
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

### Get Single Blog
```bash
GET /blogs/abc123
```

### Create Blog
```bash
POST /blogs
Content-Type: application/json

{
  "title": "New Post",
  "content": "Content here",
  "author": "John Doe",
  "tags": ["rest", "api"]
}
```

### Search
```bash
GET /search?q=graphql
```

## GraphQL API

### Endpoint
```
POST /graphql
Content-Type: application/json
```

### Queries

**Get only titles (minimal data):**
```graphql
{
  blogs {
    title
  }
}
```

**Get filtered by author:**
```graphql
{
  blogs(author: "Alice Johnson") {
    title
    content
    tags
  }
}
```

**Get specific blog:**
```graphql
{
  blog(id: "abc123") {
    title
    author
    created_at
  }
}
```

**Full blog data:**
```graphql
{
  blogs {
    id
    title
    content
    author
    tags
    created_at
  }
}
```

### Mutations

**Create blog:**
```graphql
mutation {
  createBlog(
    title: "GraphQL is Awesome"
    content: "Full content here"
    author: "John Doe"
    tags: ["graphql", "demo"]
  ) {
    id
    title
    author
    created_at
  }
}
```

## Comparison Example

### REST Request
```bash
GET /blogs
```

**Response size:** ~2.5 KB (all fields for all blogs)

### GraphQL Request
```graphql
{
  blogs {
    title
    author
  }
}
```

**Response size:** ~400 bytes (only requested fields)

**Result:** GraphQL uses ~6x less bandwidth for this query.

## Code Structure

```
cmd/server/main.go          # Entry point
internal/
├── model/blog.go            # Domain models
├── repository/blog_repo.go  # In-memory storage
├── service/blog_service.go  # Business logic
└── handler/rest_handler.go  # REST HTTP handlers
pkg/graphql/
├── schema.go                # GraphQL types & resolvers
└── resolver.go              # Query/Mutation resolvers
web/index.html               # Demo UI
```

## When to Use What?

### Use REST when:
- Simple CRUD operations
- Public APIs with caching needs
- File uploads/downloads
- Teams familiar with REST conventions

### Use GraphQL when:
- Mobile apps (reduce payload size)
- Complex data relationships
- Frontend needs flexibility
- Multiple clients with different needs

## Features Demonstrated

| Feature | REST | GraphQL |
|---------|------|---------|
| List all | ✅ GET /blogs | ✅ blogs query |
| Get one | ✅ GET /blogs/{id} | ✅ blog(id) query |
| Create | ✅ POST /blogs | ✅ createBlog mutation |
| Filter | ✅ Query params | ✅ Arguments |
| Search | ✅ /search endpoint | ✅ Query argument |
| Selective fields | ❌ No | ✅ Request only needed fields |
| Single endpoint | ❌ Multiple | ✅ /graphql only |

## Running Tests

### REST API
```bash
# List all
curl http://localhost:8080/blogs

# Filter by author
curl "http://localhost:8080/blogs?author=Alice%20Johnson"

# Create blog
curl -X POST http://localhost:8080/blogs \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","content":"Body","author":"Me","tags":["test"]}'
```

### GraphQL API
```bash
# Query
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "{ blogs { title author } }"}'

# Mutation
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "mutation { createBlog(title: \"New\", content: \"Content\", author: \"Me\", tags: [\"test\"]) { id title } }"}'
```

## Resources

- [GraphQL](https://graphql.org/)
- [graphql-go](https://github.com/graphql-go/graphql)
- [REST vs GraphQL](https://www.apollographql.com/blog/graphql-vs-rest-introduction/)

## License

MIT