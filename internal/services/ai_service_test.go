package services_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/joho/godotenv"
	"github.com/sethum-VS/my-portfolio/internal/config"
	"github.com/sethum-VS/my-portfolio/internal/services"
)

func TestParseReadmeToProductContext(t *testing.T) {
	// Load the .env file from the project root
	cwd, _ := os.Getwd()
	envPath := filepath.Join(cwd, "..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Logf("Warning: Could not load .env file from %s: %v", envPath, err)
	}
	config.Load()

	readme := `# Awesome State Manager
This project solves the complex challenge of managing state in a distributed system. 

## Challenge
Managing state across microservices often leads to data inconsistency and high latency. Our goal was to create a unified state management layer that is fast and reliable.

## Solution
We built a custom state management layer using Redis and Go, which handles synchronization asynchronously and uses optimistic locking to prevent conflicts.

## Architecture
We use a microservices architecture. 

` + "```mermaid\ngraph TD;\n    A[Frontend] --> B[API Gateway];\n    B --> C[Auth Service];\n    B --> D[State Service];\n    D --> E[(Redis)];\n```" + `

## Features
- Real-time sync
- Distributed locking
- High availability

## Tech Stack
- Go
- Redis
- React
- Docker
- Kubernetes
- Next.js

![Hero Image](https://example.com/hero.webp)
`
	ctx := context.Background()
	product, err := services.ParseReadmeToProductContext(ctx, readme)

	if err != nil {
		t.Fatalf("ParseReadmeToProductContext failed: %v", err)
	}

	if product == nil {
		t.Fatalf("Expected product, got nil")
	}

	// Output some fields for manual verification
	t.Logf("Parsed Product Title: %s", product.Title)
	t.Logf("Parsed Product Subtitle: %s", product.Subtitle)
	t.Logf("Parsed Product Tech Stack: %v", product.TechStack)
	t.Logf("Parsed Product Display Stack: %v", product.DisplayStack)
	t.Logf("Parsed Product Hero GIF: %s", product.HeroGIF)
	t.Logf("Parsed Product Core Features: %v", product.KeyFeatures)

	if !strings.Contains(strings.ToLower(product.Title), "awesome state manager") {
		t.Errorf("Expected title to contain 'awesome state manager', got '%s'", product.Title)
	}

	if len(product.TechStack) == 0 {
		t.Errorf("Expected tech stack to be populated, got empty")
	}

	if product.HeroGIF == "" {
		t.Errorf("Expected hero gif to be populated, got empty")
	}
}
func TestParseReadmeToProductContext_Failover(t *testing.T) {
	// We want to test that a 400 Bad Request (like a prompt being too long)
	// does NOT trigger a failover, while a 429 Too Many Requests DOES trigger a failover.

	var primaryCalled, backupCalled bool

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch auth := r.Header.Get("Authorization"); auth {
		case "Bearer primary-key":
			primaryCalled = true
			// We simulate a 429 Too Many Requests to test the failover
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "Rate limit exceeded", "type": "requests", "param": null, "code": "rate_limit_exceeded"}}`))
		case "Bearer backup-key":
			backupCalled = true
			w.WriteHeader(http.StatusOK)
			// Return a minimal valid response
			w.Write([]byte(`{"choices": [{"message": {"content": "{\"title\":\"Mock Title\"}"}}]}`))
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	os.Setenv("NVIDIA_NIM_BASE_URL", ts.URL)
	os.Setenv("NVIDIA_NIM_API", "primary-key")
	os.Setenv("NVIDIA_API_KEY_BACKUP", "backup-key")
	config.Load()
	defer func() {
		os.Unsetenv("NVIDIA_NIM_BASE_URL")
		os.Unsetenv("NVIDIA_NIM_API")
		os.Unsetenv("NVIDIA_API_KEY_BACKUP")
		config.Load()
	}()

	ctx := context.Background()

	// Test 1: 429 should trigger failover
	_, err := services.ParseReadmeToProductContext(ctx, "mock readme")
	if err != nil {
		t.Fatalf("Expected success with failover on 429, got err: %v", err)
	}
	if !primaryCalled || !backupCalled {
		t.Errorf("Expected both primary and backup to be called on 429. Primary: %v, Backup: %v", primaryCalled, backupCalled)
	}

	// Reset flags
	primaryCalled = false
	backupCalled = false

	// Test 2: 400 Bad Request (lengthy prompt) should NOT trigger failover
	ts400 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer primary-key" {
			primaryCalled = true
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "Context length exceeded", "type": "invalid_request_error", "param": null, "code": "context_length_exceeded"}}`))
			return
		}
		backupCalled = true
	}))
	defer ts400.Close()

	os.Setenv("NVIDIA_NIM_BASE_URL", ts400.URL)
	config.Load()
	_, err = services.ParseReadmeToProductContext(ctx, "mock lengthy prompt")

	if err == nil {
		t.Fatalf("Expected error on 400 Bad Request, got success")
	}
	if !primaryCalled {
		t.Errorf("Expected primary to be called on 400")
	}
	if backupCalled {
		t.Errorf("Expected backup NOT to be called on 400, but it was")
	}
}
