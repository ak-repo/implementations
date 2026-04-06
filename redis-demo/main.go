// Redis Demo - All use cases in ONE file
// Demonstrates: KV, TTL, Cache, Rate Limit, Lock, Pub/Sub, Queue, Counter, Session

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// ========================
// Redis Client Init
// ========================

var rdb *redis.Client
var ctx = context.Background()

func initRedis() error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}
	return nil
}

// ========================
// Utility Functions
// ========================

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ========================
// Handlers: Basic KV
// ========================

func handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")

	if key == "" || value == "" {
		respondError(w, http.StatusBadRequest, "key and value required")
		return
	}

	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"key":   key,
		"value": value,
		"ok":    "true",
	})
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "key required")
		return
	}

	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		respondJSON(w, http.StatusOK, map[string]string{
			"key":   key,
			"value": "",
			"found": "false",
		})
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"key":   key,
		"value": val,
		"found": "true",
	})
}

// ========================
// Handlers: TTL
// ========================

func handleSetTTL(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	value := r.URL.Query().Get("value")
	ttlStr := r.URL.Query().Get("ttl")

	if key == "" || value == "" || ttlStr == "" {
		respondError(w, http.StatusBadRequest, "key, value, and ttl required")
		return
	}

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "ttl must be a number")
		return
	}

	err = rdb.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"key":   key,
		"value": value,
		"ttl":   ttl,
		"ok":    true,
	})
}

// ========================
// Handlers: Cache
// ========================

type cacheData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

func handleCache(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		key = "user:1"
	}

	start := time.Now()

	// Try cache first
	cached, err := rdb.Get(ctx, key).Result()
	if err == nil {
		// Cache hit
		elapsed := time.Since(start).Milliseconds()
		var data cacheData
		json.Unmarshal([]byte(cached), &data)

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"data":    data,
			"source":  "cache",
			"time_ms": elapsed,
		})
		return
	}

	// Cache miss - simulate DB
	time.Sleep(500 * time.Millisecond)

	// Generate fake data
	data := cacheData{
		ID:        key,
		Name:      "John Doe",
		Email:     "john@example.com",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	jsonData, _ := json.Marshal(data)
	rdb.Set(ctx, key, jsonData, 60*time.Second)

	elapsed := time.Since(start).Milliseconds()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":    data,
		"source":  "db",
		"time_ms": elapsed,
	})
}

// ========================
// Handlers: Rate Limiter
// ========================

func handleRateLimit(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "anonymous"
	}

	key := fmt.Sprintf("rate_limit:%s", user)

	// Get current count
	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Set expiry on first request
	if count == 1 {
		rdb.Expire(ctx, key, time.Minute)
	}

	// Check limit (5 per minute)
	allowed := count <= 5
	status := "allowed"
	if !allowed {
		status = "blocked"
	}

	// Get remaining time
	ttl, _ := rdb.TTL(ctx, key).Result()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user":         user,
		"allowed":      allowed,
		"count":        count,
		"limit":        5,
		"status":       status,
		"reset_in_sec": int(ttl.Seconds()),
	})
}

// ========================
// Handlers: Distributed Lock
// ========================

func handleLock(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		resource = "default"
	}

	key := fmt.Sprintf("lock:%s", resource)
	lockValue := generateSessionID()

	// Try to acquire lock (SET NX PX)
	acquired, err := rdb.SetNX(ctx, key, lockValue, 5*time.Second).Result()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if acquired {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"acquired": true,
			"resource": resource,
			"lock_id":  lockValue,
			"expires":  "5 seconds",
		})
	} else {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"acquired": false,
			"resource": resource,
			"message":  "lock already held",
		})
	}
}

// ========================
// Handlers: Pub/Sub
// ========================

func handlePublish(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	if msg == "" {
		msg = "Hello from Redis!"
	}

	channel := "demo-channel"
	err := rdb.Publish(ctx, channel, msg).Err()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"published": true,
		"channel":   channel,
		"message":   msg,
	})
}

// Start subscriber goroutine
func startSubscriber() {
	go func() {
		pubsub := rdb.Subscribe(ctx, "demo-channel")
		defer pubsub.Close()

		log.Println("[Pub/Sub] Subscriber started on 'demo-channel'")

		ch := pubsub.Channel()
		for msg := range ch {
			log.Printf("[Pub/Sub] Received: %s", msg.Payload)
		}
	}()
}

// ========================
// Handlers: Queue Worker
// ========================

type Job struct {
	ID   string `json:"id"`
	Task string `json:"task"`
	Time string `json:"time"`
}

var jobHistory []Job
var jobHistoryMu sync.Mutex

type historyResponse struct {
	History []Job `json:"history"`
}

func handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondError(w, http.StatusMethodNotAllowed, "POST required")
		return
	}

	var job Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if job.ID == "" {
		job.ID = generateSessionID()
	}
	if job.Task == "" {
		job.Task = "default-task"
	}
	job.Time = time.Now().Format(time.RFC3339)

	jobJSON, _ := json.Marshal(job)
	err := rdb.LPush(ctx, "jobs", jobJSON).Err()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"queued": true,
		"job_id": job.ID,
		"task":   job.Task,
		"queue":  "jobs",
	})
}

func handleQueueHistory(w http.ResponseWriter, r *http.Request) {
	jobHistoryMu.Lock()
	history := make([]Job, len(jobHistory))
	copy(history, jobHistory)
	jobHistoryMu.Unlock()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"history": history,
	})
}

// Start worker goroutine
func startWorker() {
	go func() {
		log.Println("[Worker] Started, waiting for jobs...")

		for {
			// BRPOP blocks until job available
			result, err := rdb.BRPop(ctx, 0*time.Second, "jobs").Result()
			if err != nil {
				log.Printf("[Worker] Error: %v", err)
				continue
			}

			if len(result) < 2 {
				continue
			}

			jobJSON := result[1]
			var job Job
			json.Unmarshal([]byte(jobJSON), &job)

			log.Printf("[Worker] Processing job: %s (%s)", job.ID, job.Task)
			time.Sleep(1 * time.Second) // Simulate work
			log.Printf("[Worker] Completed job: %s", job.ID)

			// Add to history
			jobHistoryMu.Lock()
			jobHistory = append([]Job{job}, jobHistory...)
			if len(jobHistory) > 20 {
				jobHistory = jobHistory[:20]
			}
			jobHistoryMu.Unlock()
		}
	}()
}

// ========================
// Handlers: Counter
// ========================

func handleCounter(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		key = "counter"
	}

	count, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"key":   key,
		"count": count,
	})
}

// ========================
// Handlers: Session
// ========================

func handleLogin(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		user = "guest"
	}

	sessionID := generateSessionID()
	sessionKey := fmt.Sprintf("session:%s", sessionID)

	err := rdb.Set(ctx, sessionKey, user, 5*time.Minute).Err()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": sessionID,
		"user":       user,
		"expires_in": "5 minutes",
	})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		respondError(w, http.StatusUnauthorized, "session_id required")
		return
	}

	sessionKey := fmt.Sprintf("session:%s", sessionID)
	user, err := rdb.Get(ctx, sessionKey).Result()

	if err == redis.Nil {
		respondError(w, http.StatusUnauthorized, "session expired or invalid")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get remaining TTL
	ttl, _ := rdb.TTL(ctx, sessionKey).Result()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user":           user,
		"session_id":     sessionID,
		"expires_in_sec": int(ttl.Seconds()),
	})
}

// ========================
// main()
// ========================

func main() {
	// Connect to Redis
	if err := initRedis(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis!")

	// Start background workers
	startSubscriber()
	startWorker()

	// Setup routes
	mux := http.NewServeMux()

	// KV
	mux.HandleFunc("/set", handleSet)
	mux.HandleFunc("/get", handleGet)

	// TTL
	mux.HandleFunc("/set-ttl", handleSetTTL)

	// Cache
	mux.HandleFunc("/cache", handleCache)

	// Rate Limiter
	mux.HandleFunc("/rate-limit", handleRateLimit)

	// Lock
	mux.HandleFunc("/lock", handleLock)

	// Pub/Sub
	mux.HandleFunc("/publish", handlePublish)

	// Queue
	mux.HandleFunc("/enqueue", handleEnqueue)
	mux.HandleFunc("/queue/history", handleQueueHistory)

	// Counter
	mux.HandleFunc("/counter", handleCounter)

	// Session
	mux.HandleFunc("/login", handleLogin)
	mux.HandleFunc("/me", handleMe)

	// Health
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Static files
	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// Apply CORS
	handler := enableCORS(mux)

	// Logging middleware
	loggedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		handler.ServeHTTP(w, r)
	})

	// Start server
	log.Println("Server starting on http://localhost:8080")
	log.Println("Web UI: http://localhost:8080/index.html")
	log.Fatal(http.ListenAndServe(":8080", loggedHandler))
}
