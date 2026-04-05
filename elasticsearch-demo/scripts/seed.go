// Package main generates sample blog data for Elasticsearch.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	// Sample data
	authors := []string{
		"Alex Johnson",
		"Sarah Chen",
		"Mike Rivera",
		"Emily Zhang",
		"David Kim",
		"Lisa Thompson",
	}

	tags := [][]string{
		{"golang", "backend"},
		{"javascript", "frontend", "react"},
		{"docker", "devops", "kubernetes"},
		{"database", "postgresql", "performance"},
		{"microservices", "architecture", "api"},
		{"security", "authentication", "oauth"},
		{"testing", "unit-tests", "tdd"},
		{"cloud", "aws", "serverless"},
		{"monitoring", "observability", "prometheus"},
		{"ci-cd", "github-actions", "automation"},
	}

	titles := []string{
		"Getting Started with Go Generics",
		"Building Scalable Microservices with Docker and Kubernetes",
		"Understanding PostgreSQL Indexing Strategies",
		"Modern React Patterns: Hooks and Context",
		"Securing Your API with JWT and OAuth 2.0",
		"Test-Driven Development in Practice",
		"Deploying Serverless Functions on AWS Lambda",
		"Monitoring Production Systems with Prometheus",
		"Building CI/CD Pipelines with GitHub Actions",
		"Database Performance Optimization Tips",
		"Introduction to gRPC for Microservices",
		"Handling Concurrency in Go: Best Practices",
		"REST API Design Principles",
		"GraphQL vs REST: Making the Right Choice",
		"Setting Up Elasticsearch for Full-Text Search",
		"Caching Strategies for Web Applications",
		"Event-Driven Architecture Explained",
		"Zero-Downtime Deployments with Kubernetes",
		"Authentication Patterns in Modern Web Apps",
		"Building Real-Time Applications with WebSockets",
		"Docker Multi-Stage Builds for Production",
		"Effective Logging in Distributed Systems",
		"Load Balancing Strategies for High Traffic",
		"API Rate Limiting: Implementation Guide",
		"Database Sharding: When and How",
		"Introduction to Service Mesh with Istio",
		"Web Security: Common Vulnerabilities and Fixes",
		"Building a Blog Platform with Go and Elasticsearch",
		"Optimizing JavaScript Bundle Size",
		"Continuous Integration Best Practices",
		"Data Migration Strategies for Production",
		"Implementing Search Autocomplete",
		"Using Redis for Session Management",
		"Message Queue Patterns with RabbitMQ",
		"Creating Custom GitHub Actions",
		"Understanding HTTP/2 and Server Push",
		"Practical Guide to Go Interfaces",
		"Error Handling in Distributed Systems",
		"Building Responsive UIs with Tailwind CSS",
		"Introduction to Infrastructure as Code",
		"Deep Dive into Go Channels",
		"API Versioning Strategies",
		"Database Connection Pooling Explained",
		"Container Orchestration at Scale",
		"Optimizing API Latency",
		"Building Command Line Tools in Go",
		"Understanding CORS and How to Fix It",
		"WebSocket Scaling Patterns",
		"Implementing Circuit Breakers",
		"Distributed Tracing with Jaeger",
		"Secrets Management in Kubernetes",
		"Writing Maintainable Go Code",
		"Database Migration Best Practices",
		"Introduction to eBPF",
		"Building a Task Queue with Redis",
		"Understanding TCP and HTTP",
		"Modern CSS Grid Techniques",
		"Rate Limiting Algorithms Explained",
		"Service Discovery Patterns",
		"Optimizing Go Memory Usage",
		"Building Server-Side Rendered Apps",
		"Understanding Event Sourcing",
		"PostgreSQL Query Optimization",
		"Introduction to WASM",
		"Building Multi-Tenant Applications",
		"CI/CD Security Best Practices",
		"Go Profiling and Optimization",
		"Working with Time Zones in Applications",
		"Building Analytics Pipelines",
		"Understanding SQL Transactions",
		"Real-Time Data Processing with Kafka",
		"Building Plugin Architectures",
		"API Documentation with OpenAPI",
		"Go Context Package Deep Dive",
		"Managing Technical Debt",
		"Building Notification Systems",
		"Understanding Linux Namespaces",
		"Graph Database Fundamentals",
		"Go Testing Patterns",
		"Building Feature Flags Systems",
		"Introduction to Chaos Engineering",
		"Working with Protocol Buffers",
		"Building Recommendation Engines",
		"Understanding OAuth Flows",
		"Go Memory Model Explained",
		"Building Multi-Region Systems",
		"Database Replication Strategies",
		"Introduction to Vector Databases",
	}

	contentTemplates := []string{
		"In this comprehensive guide, we explore %s and its practical applications in modern software development. We'll cover best practices, common pitfalls, and real-world examples to help you master this concept.",
		"Learn how to implement %s effectively in your projects. This tutorial walks through the setup, configuration, and optimization techniques you need to know.",
		"What makes %s essential for developers today? We dive deep into the core concepts, explore various use cases, and provide actionable insights.",
		"This article examines %s from multiple angles. Whether you're a beginner or experienced developer, you'll find valuable insights and practical tips.",
		"Discover the power of %s in this hands-on tutorial. We build a complete example from scratch, explaining each step along the way.",
		"Mastering %s can significantly improve your development workflow. Here's everything you need to know to get started.",
		"Let's demystify %s. We'll break down complex concepts into simple terms and show you how to apply them in real projects.",
		"Exploring %s opens up new possibilities for your applications. This guide covers the fundamentals and advanced techniques.",
		"When should you use %s? This article helps you understand the trade-offs and make informed architectural decisions.",
		"A deep dive into %s - from theory to practice. We examine implementation details and share lessons learned from production systems.",
	}

	// Seed random
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Seeding Elasticsearch with blog data...")
	fmt.Println()

	baseURL := "http://localhost:8080"

	// Check if server is running
	_, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("Error: Cannot connect to server at %s\n", baseURL)
		fmt.Println("Make sure the server is running: go run cmd/server/main.go")
		return
	}

	successCount := 0

	for i, title := range titles {
		// Generate content
		content := fmt.Sprintf(contentTemplates[i%len(contentTemplates)], title)
		content += generateParagraph()

		// Select random author and tags
		author := authors[rand.Intn(len(authors))]
		tagSet := tags[rand.Intn(len(tags))]

		// Create blog
		blog := map[string]interface{}{
			"title":   title,
			"content": content,
			"author":  author,
			"tags":    tagSet,
		}

		jsonData, _ := json.Marshal(blog)

		resp, err := http.Post(baseURL+"/blogs", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create blog '%s': %v\n", title, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusCreated {
			successCount++
			fmt.Printf("✓ Created: %s\n", title)
		} else {
			fmt.Printf("✗ Failed: %s (status: %d)\n", title, resp.StatusCode)
		}

		// Small delay to avoid overwhelming the server
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println()
	fmt.Printf("Successfully seeded %d/%d blog posts\n", successCount, len(titles))
	fmt.Println()
	fmt.Println("You can now search at: http://localhost:8080")
}

func generateParagraph() string {
	paragraphs := []string{
		" This approach has been battle-tested in production environments and has proven to significantly improve performance and maintainability.",
		" Many developers find this technique particularly useful when scaling applications to handle millions of requests.",
		" Understanding these concepts is crucial for building robust and scalable systems that can handle real-world traffic.",
		" We'll also discuss common anti-patterns and how to avoid them in your own implementations.",
		" By the end of this article, you'll have a solid foundation to implement this in your own projects.",
		" The examples provided are production-ready and follow industry best practices.",
		" This pattern is widely used by companies like Google, Netflix, and Uber to handle massive scale.",
		" We'll explore both the theoretical foundations and practical implementation details.",
		" Security considerations are also covered to ensure your implementation is robust against common attacks.",
		" Performance benchmarks show significant improvements when applying these techniques correctly.",
	}

	result := ""
	for i := 0; i < 3+rand.Intn(3); i++ {
		result += paragraphs[rand.Intn(len(paragraphs))]
	}
	return result
}
