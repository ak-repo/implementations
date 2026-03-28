package main

import (
	"fmt"
	"log"
	"net/http"

	"s3-upload-demo/internal/config"
	"s3-upload-demo/internal/handler"
	"s3-upload-demo/internal/s3"
	"s3-upload-demo/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	s3Client, err := s3.NewClient(
		cfg.AWSAccessKeyID,
		cfg.AWSSecretAccessKey,
		cfg.AWSRegion,
		cfg.S3BucketName,
	)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	uploadService := service.NewUploadService(s3Client)
	uploadHandler := handler.NewUploadHandler(uploadService)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/upload/generate-url", uploadHandler.GenerateUploadURL)
	mux.HandleFunc("/api/v1/upload/file", uploadHandler.UploadFile)
	mux.Handle("/", http.FileServer(http.Dir("frontend")))

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	corsHandler := withCORS(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("API endpoint: POST http://localhost%s/api/v1/upload/generate-url", addr)

	if err := http.ListenAndServe(addr, corsHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
