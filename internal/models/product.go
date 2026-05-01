package models

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// DB holds the Firestore client for the models package.
// It should be initialized from main.go
var DB *firestore.Client

// InitDB sets the Firestore client.
func InitDB(client *firestore.Client) {
	DB = client
}

// Product represents a portfolio project with detailed engineering and architectural information.
type Product struct {
	ID           string            `firestore:"id" json:"id"`
	Title        string            `firestore:"title" json:"title"`
	Subtitle     string            `firestore:"subtitle" json:"subtitle"`
	Description  string            `firestore:"description" json:"description"`
	HeroGIF      string            `firestore:"hero_gif" json:"hero_gif"`
	Challenge    string            `firestore:"challenge" json:"challenge"`
	Solution     string            `firestore:"solution" json:"solution"`
	Architecture string            `firestore:"architecture" json:"architecture"`
	ArchDiagram  string            `firestore:"arch_diagram" json:"arch_diagram"`
	InternalFlow []string          `firestore:"internal_flow" json:"internal_flow"`
	TechStack    []string          `firestore:"tech_stack" json:"tech_stack"`
	DisplayStack []string          `firestore:"display_stack" json:"display_stack"`
	KeyFeatures  []string          `firestore:"key_features" json:"key_features"`
	LiveURL      string            `firestore:"live_url" json:"live_url"`
	GitHubURL    string            `firestore:"github_url" json:"github_url"`
	Metrics      map[string]string `firestore:"metrics" json:"metrics"`
	Deployment   string            `firestore:"deployment" json:"deployment"`
}

// AllProducts returns the catalog of all portfolio projects from Firestore.
func AllProducts() []Product {
	ctx := context.Background()
	var products []Product

	if DB == nil {
		log.Println("Firestore client not initialized")
		return products
	}

	iter := DB.Collection("projects").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate over projects: %v", err)
			return products
		}
		var p Product
		if err := doc.DataTo(&p); err != nil {
			log.Printf("Failed to map document to struct: %v", err)
			continue
		}
		products = append(products, p)
	}

	return products
}

// GetProductByID finds and returns a product by its ID from Firestore.
func GetProductByID(id string) *Product {
	ctx := context.Background()

	if DB == nil {
		log.Println("Firestore client not initialized")
		return nil
	}

	doc, err := DB.Collection("projects").Doc(id).Get(ctx)
	if err != nil {
		// Document doesn't exist or other error
		return nil
	}

	var p Product
	if err := doc.DataTo(&p); err != nil {
		log.Printf("Failed to map document to struct: %v", err)
		return nil
	}

	return &p
}

// CreateProduct adds a new product to Firestore.
func CreateProduct(p Product) error {
	ctx := context.Background()

	if DB == nil {
		return fmt.Errorf("Firestore client not initialized")
	}

	// Check if exists first
	doc, err := DB.Collection("projects").Doc(p.ID).Get(ctx)
	if err == nil && doc.Exists() {
		return fmt.Errorf("product with ID %s already exists", p.ID)
	}

	_, err = DB.Collection("projects").Doc(p.ID).Set(ctx, p)
	if err != nil {
		return fmt.Errorf("failed to create product: %v", err)
	}

	return nil
}

// UpdateProduct modifies an existing product in Firestore.
func UpdateProduct(id string, updated Product) error {
	ctx := context.Background()

	if DB == nil {
		return fmt.Errorf("Firestore client not initialized")
	}

	// If ID changed, we need to create a new doc and delete the old one
	if id != updated.ID {
		err := CreateProduct(updated)
		if err != nil {
			return err
		}
		return DeleteProduct(id)
	}

	// Otherwise just set (overwrite) the existing document
	_, err := DB.Collection("projects").Doc(id).Set(ctx, updated)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	return nil
}

// DeleteProduct removes a product from Firestore.
func DeleteProduct(id string) error {
	ctx := context.Background()

	if DB == nil {
		return fmt.Errorf("Firestore client not initialized")
	}

	_, err := DB.Collection("projects").Doc(id).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}

	return nil
}
