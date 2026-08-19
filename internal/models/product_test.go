package models

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestAllProducts(t *testing.T) {
	_ = godotenv.Load("../../.env")
	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		t.Skip("SUPABASE_DB_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	InitDB(pool)
	invalidateProductCache()

	products := AllProducts(ctx)
	t.Logf("AllProducts returned: %d products", len(products))
	for _, p := range products {
		t.Logf("- %s: %s", p.ID, p.Title)
	}

	// Test raw scan
	query := `SELECT id, title, subtitle, description, hero_gif, challenge, solution, architecture, arch_diagram, internal_flow, tech_stack, display_stack, key_features, live_url, github_url, metrics, deployment FROM projects`
	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	i := 0
	for rows.Next() {
		i++
		var p Product
		err := rows.Scan(
			&p.ID, &p.Title, &p.Subtitle, &p.Description, &p.HeroGIF, &p.Challenge, &p.Solution, &p.Architecture, &p.ArchDiagram,
			&p.InternalFlow, &p.TechStack, &p.DisplayStack, &p.KeyFeatures, &p.LiveURL, &p.GitHubURL, &p.Metrics, &p.Deployment,
		)
		if err != nil {
			t.Errorf("Row %d Scan Error: %v", i, err)
		} else {
			t.Logf("Row %d Scanned Successfully: %s", i, p.Title)
		}
	}
}
