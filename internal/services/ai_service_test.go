package services_test

import (
	"context"
	"strings"
	"testing"
	"path/filepath"
	"os"

	"github.com/joho/godotenv"
	"github.com/sethum-VS/my-portfolio/internal/services"
)

func TestParseReadmeToProductContext(t *testing.T) {
	// Load the .env file from the project root
	cwd, _ := os.Getwd()
	envPath := filepath.Join(cwd, "..", "..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		t.Logf("Warning: Could not load .env file from %s: %v", envPath, err)
	}

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
