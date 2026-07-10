package services_test

import (
	"context"
	"strings"
	"testing"

	"github.com/sethum-VS/my-portfolio/internal/services"
)

func TestParseReadmeToProductContext(t *testing.T) {
	readme := `# Test Project
This is a test project.
## Features
- Feature 1
- Feature 2
`
	ctx := context.Background()
	product, err := services.ParseReadmeToProductContext(ctx, readme)

	if err != nil {
		t.Fatalf("ParseReadmeToProductContext failed: %v", err)
	}

	if product == nil {
		t.Fatalf("Expected product, got nil")
	}

	if !strings.Contains(product.Title, "Test") {
		t.Errorf("Expected title to contain Test, got %s", product.Title)
	}
}
