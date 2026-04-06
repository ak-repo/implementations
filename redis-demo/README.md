# Redis Demo

A production-style demo showcasing real-world Redis use cases with Go.

## What is Redis?

Redis (Remote Dictionary Server) is an **in-memory data structure store** used as:
- Database
- Cache
- Message broker
- Streaming engine

**Why Redis?**
- ⚡ **Speed**: Sub-millisecond latency (in-memory)
- 🔄 **Versatility**: Strings, Lists, Sets, Sorted Sets, Hashes, Streams
- 📊 **Persistence**: Optional disk persistence
- 🚀 **Scalability**: Replication, Clustering, Partitioning

## Features Demonstrated

| Feature | Use Case | Redis Commands |
|---------|----------|----------------|
| **Cache** | Speed up DB queries | `GET`, `SET`, `EXPIRE` |
| **Rate Limit** | Prevent API abuse | `INCR`, `EXPIRE`, `TTL` |
| **Distributed Lock** | Coordinate distributed systems | `SET NX PX` |
| **Pub/Sub** | Real-time messaging | `PUBLISH`, `SUBSCRIBE` |
| **Queue** | Background job processing | `LPUSH`, `BRPOP` |
| **Counter** | Real-time statistics | `INCR`, `DECR` |
| **Session** | User authentication state | `SET`, `GET`, `EXPIRE` |

## Quick Start

### 1. Start Redis

```bash
cd redis-demo
docker-compose up -d
```

Verify Redis is ready:
```bash
redis-cli ping
# Should return: PONG
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Run the Server

```bash
go run main.go
```

Server starts on `http://localhost:8080`

### 4. Open Web UI

```bash
open http://localhost:8080
```

## API Endpoints

### Basic KV
```bash
# Set a key
GET /set?key=foo&value=bar

# Get a key
GET /get?key=foo
```

### TTL (Expiration)
```bash
# Set with 10 second TTL
GET /set-ttl?key=temp&value=data&ttl=10

# Key auto-expires after 10 seconds
```

### Cache Demo
```bash
# First call: slow (simulates DB, 500ms)
# Subsequent calls: fast (from Redis cache)
GET /cache?key=user:1

Response:
{
  "data": { "id": "user:1", "name": "John Doe", ... },
  "source": "cache",  // or "db" on first call
  "time_ms": 5
}
```

### Rate Limiter
```bash
# 5 requests per minute allowed
GET /rate-limit?user=user123

Response:
{
  "user": "user123",
  "allowed": true,
  "count": 3,
  "limit": 5,
  "status": "allowed",
  "reset_in_sec": 45
}
```

### Distributed Lock
```bash
# Try to acquire lock on resource
GET /lock?resource=job1

Response (acquired):
{
  "acquired": true,
  "resource": "job1",
  "lock_id": "abc123...",
  "expires": "5 seconds"
}

Response (already locked):
{
  "acquired": false,
  "resource": "job1",
  "message": "lock already held"
}
```

### Pub/Sub
```bash
# Publish message (subscriber runs in background)
GET /publish?msg=Hello%20Redis

# Check server logs to see subscriber receiving messages
```

### Queue Worker
```bash
# Add job to queue
POST /enqueue
Content-Type: application/json

{
  "id": "job-123",
  "task": "send-email"
}

# Worker processes jobs in background (1 second delay simulated)
# Check server logs or GET /queue/history
```

### Counter
```bash
# Increment counter
GET /counter?key=views

Response:
{
  "key": "views",
  "count": 42
}
```

### Session
```bash
# Create session (5 min TTL)
GET /login?user=john

Response:
{
  "session_id": "abc123...",
  "user": "john",
  "expires_in": "5 minutes"
}

# Verify session
GET /me?session_id=abc123...

Response:
{
  "user": "john",
  "session_id": "abc123...",
  "expires_in_sec": 287
}
```

## Architecture

```
┌─────────────────┐
│   Web Client    │
│  (index.html)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌─────────────────┐
│   Go Server     │────>│     Redis       │
│   (main.go)     │     │   Port 6379     │
│                 │     │                 │
│ • KV Handlers   │     │ • Cache         │
│ • Cache Pattern │     │ • Rate Limit    │
│ • Rate Limiter  │     │ • Pub/Sub       │
│ • Queue Worker  │     │ • Queue (List)  │
│ • Pub/Sub Sub   │     │ • Session Store │
└─────────────────┘     └─────────────────┘
```

## Code Structure

```go
main.go (single file, ~650 lines)
├── Redis Client Init
├── Utility Functions
├── Handlers (9 use cases)
│   ├── Basic KV (/set, /get)
│   ├── TTL (/set-ttl)
│   ├── Cache (/cache)
│   ├── Rate Limit (/rate-limit)
│   ├── Distributed Lock (/lock)
│   ├── Pub/Sub (/publish + subscriber)
│   ├── Queue (/enqueue + worker)
│   ├── Counter (/counter)
│   └── Session (/login, /me)
└── main() - Routes + Server
```

## Testing

### Cache Pattern
```bash
# First call - hits DB (slow)
curl "http://localhost:8080/cache?key=user:1"

# Second call - hits Redis (fast)
curl "http://localhost:8080/cache?key=user:1"
```

### Rate Limiting
```bash
# Run 7 times - first 5 allowed, 6th+ blocked
for i in {1..7}; do
  curl "http://localhost:8080/rate-limit?user=test"
done
```

### Queue Worker
```bash
# Add multiple jobs
curl -X POST http://localhost:8080/enqueue \
  -H "Content-Type: application/json" \
  -d '{"task":"send-welcome-email"}'

# Watch server logs to see worker processing
```

## Monitoring

### Check Redis Keys
```bash
redis-cli

# List all keys
KEYS *

# Check cache
GET user:1

# Check rate limit
GET rate_limit:user123
TTL rate_limit:user123

# Check session
KEYS session:*

# Check queue length
LLEN jobs

# Check counter
GET views
```

## Cleanup

```bash
# Stop containers
docker-compose down

# Remove data
docker-compose down -v
```

## Resources

- [Redis Commands](https://redis.io/commands)
- [Go Redis Client](https://github.com/redis/go-redis)
- [Redis Patterns](https://redis.io/docs/manual/patterns/)

## License

MIT