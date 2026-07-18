package models

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the PostgreSQL connection pool for the models package.
// It should be initialized from main.go
var DB *pgxpool.Pool

// InitDB sets the DB pool.
func InitDB(pool *pgxpool.Pool) {
	DB = pool
}

func checkDB() error {
	if DB == nil {
		log.Println("DB pool not initialized")
		return fmt.Errorf("DB pool not initialized")
	}
	return nil
}

// ── TTL Cache for AllProducts ────────────────────────────────────────────────
var (
	productCacheMu   sync.RWMutex
	productCache     []Product
	productCacheTime time.Time
	productCacheTTL  = 60 * time.Second
)

// invalidateProductCache clears the cached product list.
func invalidateProductCache() {
	productCacheMu.Lock()
	productCache = nil
	productCacheTime = time.Time{}
	productCacheMu.Unlock()
}

// Product represents a portfolio project.
type Product struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Subtitle     string            `json:"subtitle"`
	Description  string            `json:"description"`
	HeroGIF      string            `json:"hero_gif"`
	Challenge    string            `json:"challenge"`
	Solution     string            `json:"solution"`
	Architecture string            `json:"architecture"`
	ArchDiagram  string            `json:"arch_diagram"`
	InternalFlow []string          `json:"internal_flow"`
	TechStack    []string          `json:"tech_stack"`
	DisplayStack []string          `json:"display_stack"`
	KeyFeatures  []string          `json:"key_features"`
	LiveURL      string            `json:"live_url"`
	GitHubURL    string            `json:"github_url"`
	Metrics      map[string]string `json:"metrics"`
	Deployment   string            `json:"deployment"`
}

// AllProducts returns the catalog of all portfolio projects.
func AllProducts(ctx context.Context) []Product {
	// Check cache first
	productCacheMu.RLock()
	if productCache != nil && time.Since(productCacheTime) < productCacheTTL {
		cached := make([]Product, len(productCache))
		copy(cached, productCache)
		productCacheMu.RUnlock()
		return cached
	}
	productCacheMu.RUnlock()

	var products []Product

	if err := checkDB(); err != nil {
		return products
	}

	query := `SELECT id, title, subtitle, description, hero_gif, challenge, solution, architecture, arch_diagram, internal_flow, tech_stack, display_stack, key_features, live_url, github_url, metrics, deployment FROM projects`
	rows, err := DB.Query(ctx, query)
	if err != nil {
		log.Printf("Failed to execute query: %v", err)
		return products
	}
	defer rows.Close()

	for rows.Next() {
		var p Product
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Subtitle, &p.Description, &p.HeroGIF, &p.Challenge, &p.Solution, &p.Architecture, &p.ArchDiagram,
			&p.InternalFlow, &p.TechStack, &p.DisplayStack, &p.KeyFeatures, &p.LiveURL, &p.GitHubURL, &p.Metrics, &p.Deployment,
		); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		products = append(products, p)
	}

	// Update cache
	productCacheMu.Lock()
	productCache = make([]Product, len(products))
	copy(productCache, products)
	productCacheTime = time.Now()
	productCacheMu.Unlock()

	return products
}

// GetProductByID finds and returns a product by its ID.
func GetProductByID(ctx context.Context, id string) *Product {
	if err := checkDB(); err != nil {
		return nil
	}

	var p Product
	query := `SELECT id, title, subtitle, description, hero_gif, challenge, solution, architecture, arch_diagram, internal_flow, tech_stack, display_stack, key_features, live_url, github_url, metrics, deployment FROM projects WHERE id = $1`
	err := DB.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Title, &p.Subtitle, &p.Description, &p.HeroGIF, &p.Challenge, &p.Solution, &p.Architecture, &p.ArchDiagram,
		&p.InternalFlow, &p.TechStack, &p.DisplayStack, &p.KeyFeatures, &p.LiveURL, &p.GitHubURL, &p.Metrics, &p.Deployment,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		log.Printf("Failed to get product: %v", err)
		return nil
	}

	return &p
}

// CreateProduct adds a new product.
func CreateProduct(ctx context.Context, p Product) error {
	if err := checkDB(); err != nil {
		return err
	}

	query := `
		INSERT INTO projects (
			id, title, subtitle, description, hero_gif, challenge, solution, architecture, arch_diagram, internal_flow, tech_stack, display_stack, key_features, live_url, github_url, metrics, deployment
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)`
	_, err := DB.Exec(ctx, query,
		p.ID, p.Title, p.Subtitle, p.Description, p.HeroGIF, p.Challenge, p.Solution, p.Architecture, p.ArchDiagram,
		p.InternalFlow, p.TechStack, p.DisplayStack, p.KeyFeatures, p.LiveURL, p.GitHubURL, p.Metrics, p.Deployment,
	)
	if err != nil {
		return fmt.Errorf("failed to create product: %v", err)
	}

	invalidateProductCache()
	return nil
}

// UpdateProduct modifies an existing product.
func UpdateProduct(ctx context.Context, id string, updated Product) error {
	if err := checkDB(); err != nil {
		return err
	}

	// If ID changed, delete old one
	if id != updated.ID {
		err := DeleteProduct(ctx, id)
		if err != nil {
			return err
		}
	}

	query := `
		INSERT INTO projects (
			id, title, subtitle, description, hero_gif, challenge, solution, architecture, arch_diagram, internal_flow, tech_stack, display_stack, key_features, live_url, github_url, metrics, deployment
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			subtitle = EXCLUDED.subtitle,
			description = EXCLUDED.description,
			hero_gif = EXCLUDED.hero_gif,
			challenge = EXCLUDED.challenge,
			solution = EXCLUDED.solution,
			architecture = EXCLUDED.architecture,
			arch_diagram = EXCLUDED.arch_diagram,
			internal_flow = EXCLUDED.internal_flow,
			tech_stack = EXCLUDED.tech_stack,
			display_stack = EXCLUDED.display_stack,
			key_features = EXCLUDED.key_features,
			live_url = EXCLUDED.live_url,
			github_url = EXCLUDED.github_url,
			metrics = EXCLUDED.metrics,
			deployment = EXCLUDED.deployment`

	_, err := DB.Exec(ctx, query,
		updated.ID, updated.Title, updated.Subtitle, updated.Description, updated.HeroGIF, updated.Challenge, updated.Solution, updated.Architecture, updated.ArchDiagram,
		updated.InternalFlow, updated.TechStack, updated.DisplayStack, updated.KeyFeatures, updated.LiveURL, updated.GitHubURL, updated.Metrics, updated.Deployment,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	invalidateProductCache()
	return nil
}

// DeleteProduct removes a product.
func DeleteProduct(ctx context.Context, id string) error {
	if err := checkDB(); err != nil {
		return err
	}

	_, err := DB.Exec(ctx, "DELETE FROM projects WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %v", err)
	}

	invalidateProductCache()
	return nil
}
