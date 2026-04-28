package models

import (
	"fmt"
	"sync"
)

// Product represents a portfolio project with detailed engineering and architectural information.
type Product struct {
	ID           string            // URL-safe identifier (e.g., "route-optimizer-engine")
	Title        string            // Project title
	Subtitle     string            // Short description
	Description  string            // Detailed project description
	HeroGIF      string            // Path to the main demo GIF
	Challenge    string            // Technical challenge or problem statement
	Solution     string            // How it was solved
	Architecture string            // Architectural overview/approach
	ArchDiagram  string            // ASCII/Text diagram
	InternalFlow []string          // Step-by-step data flow
	TechStack    []string          // Technologies used
	KeyFeatures  []string          // Main features or highlights
	LiveURL      string            // Link to live project (if applicable)
	GitHubURL    string            // Link to GitHub repository (if applicable)
	Metrics      map[string]string // Performance metrics or stats
	Deployment   string            // Deployment/CI/CD info
}

// allProducts is the catalog of all portfolio projects.
var allProducts = []Product{
	{
		ID:           "deca-node-tsp",
		Title:        "Deca-Node TSP",
		Subtitle:     "Microservices-based delivery route optimization tool",
		Description:  "A robust delivery route optimization system addressing the Traveling Salesperson Problem (TSP) using real-world OSM data. Built with a decoupled microservice architecture for high-frequency matrix calculations.",
		HeroGIF:      "/static/images/deca-node-demo.webp", // Placeholder for demo.webp
		Challenge:    "Maintaining data sovereignty and avoiding commercial mapping API costs while performing high-frequency NxN distance matrix calculations for logistics optimization.",
		Solution:     "Leveraged the OSM/GraphHopper stack to build a custom routing engine that performs geocoding, snap-to-road, and TSP solving entirely within a private infrastructure.",
		Architecture: "Microservice architecture using a Go orchestrator for smart parsing and HTML fragments (HTMX), and a Java Spring Boot engine for core routing and optimization.",
		ArchDiagram: `Browser (HTMX) ─→ Go Orchestrator (:8080) ─→ Java Routing Engine (:8081)
                   │  Smart Parsing              │  Geocoding (Photon)
                   │  HTML Fragments             │  Snap-to-Road (GraphHopper)
                   └──────────────────           │  Distance Matrix (GraphHopper)
                                                 │  TSP Solver (Jsprit)
                                                 └──────────────────────────`,
		InternalFlow: []string{
			"User enters an address or clicks map via Leaflet interface",
			"Frontend fetches geographic coordinates via Photon Geocoding API",
			"Go Orchestrator parses inputs and sends JSON payload to Java engine",
			"Java engine performs Snap-to-Road and computes Distance Matrix via GraphHopper",
			"Jsprit resolves the VRP/TSP sequence mathematically",
			"Go returns an HTMX out-of-band response to render the optimized polyline",
		},
		TechStack: []string{"Go", "Java 21", "Spring Boot", "HTMX", "Leaflet.js", "GraphHopper", "Jsprit", "Docker Compose"},
		KeyFeatures: []string{
			"TSP Optimization with Jsprit",
			"Hierarchical Road Prioritization",
			"Time-Based Penalty Mechanism",
			"Commercial-Free Mapping Stack",
			"CI/CD with GCP & Firebase",
		},
		LiveURL:   "https://deca-node.example.com",
		GitHubURL: "https://github.com/sethum-VS/Deca-Node-TSP",
		Metrics: map[string]string{
			"PBF Size":       "~150MB (Sri Lanka)",
			"Startup Time":   "30-60s (Graph Build)",
			"Routing Engine": "GraphHopper 11.0",
			"CI Pipeline":    "GitHub Actions + WIF",
		},
		Deployment: "Deployed to GCP Cloud Run (internal/public tiering) with Firebase Hosting for proxying and SSL management.",
	},
	{
		ID:           "route-optimizer-engine",
		Title:        "Route Optimizer Engine",
		Subtitle:     "High-performance pathfinding system",
		Description:  "A distributed route optimization engine built with Go that processes millions of waypoints using advanced graph algorithms. Designed for real-time logistics and fleet management systems.",
		HeroGIF:      "",
		Challenge:    "Traditional routing solutions couldn't scale beyond 100K waypoints without significant latency. We needed sub-100ms response times for production traffic.",
		Solution:     "Implemented a custom spatial indexing system with quad-tree decomposition, combined with A* pathfinding optimizations and connection pooling for PostgreSQL.",
		Architecture: "Microservice architecture with Go backend handling algorithm execution, PostgreSQL for spatial data storage, and Redis for result caching. HTMX frontend provides real-time updates via WebSocket connections.",
		TechStack:    []string{"Go", "PostgreSQL", "Redis", "HTMX", "Protocol Buffers", "Docker"},
		KeyFeatures: []string{
			"Sub-100ms pathfinding on 1M+ waypoints",
			"Real-time traffic optimization",
			"Distributed caching layer",
			"Live progress visualization",
			"Horizontal scaling support",
		},
		LiveURL:   "https://route-optimizer.example.com",
		GitHubURL: "https://github.com/sethum-VS/route-optimizer-engine",
		Metrics: map[string]string{
			"Response Time": "45ms average",
			"Throughput":    "10K+ requests/sec",
			"Uptime":        "99.95%",
			"Optimization":  "35% improvement over baseline",
		},
	},
	{
		ID:           "realtime-grid-visualizer",
		Title:        "Realtime Grid Visualizer",
		Subtitle:     "GPU-accelerated data visualization engine",
		Description:  "A high-performance WebGL-based visualization system for rendering massive datasets in real-time. Built to handle 1M+ data points with 60fps interactions and smooth animations.",
		HeroGIF:      "",
		Challenge:    "Standard Canvas/SVG rendering couldn't handle more than 50K points without dropping frames. Needed smooth 60fps performance across heterogeneous hardware.",
		Solution:     "Leveraged WebGL shaders for GPU-accelerated rendering, implemented adaptive LOD (Level of Detail) system, and optimized vertex buffer management for dynamic updates.",
		Architecture: "TypeScript frontend using Three.js abstractions over raw WebGL, with a custom shader system for advanced effects. WebWorkers handle data processing off-thread, preventing UI blocking.",
		TechStack:    []string{"TypeScript", "WebGL", "Three.js", "WebWorkers", "Tailwind CSS", "Vite"},
		KeyFeatures: []string{
			"1M+ point rendering at 60fps",
			"Smooth zoom/pan interactions",
			"Custom shader effects (heat maps, flow fields)",
			"Real-time data streaming support",
			"Cross-browser optimization",
		},
		LiveURL:   "https://grid-visualizer.example.com",
		GitHubURL: "https://github.com/sethum-VS/realtime-grid-visualizer",
		Metrics: map[string]string{
			"Render Performance":  "60fps @ 1M points",
			"Load Time":           "2.1s (Lighthouse)",
			"Memory Usage":        "180MB (1M points)",
			"Interaction Latency": "16ms",
		},
	},
	{
		ID:           "portfolio-control-panel",
		Title:        "Portfolio Control Panel",
		Subtitle:     "Full-stack portfolio management system",
		Description:  "A comprehensive admin dashboard for managing portfolio content, project details, and deployment pipeline. Built as a self-contained Go application with containerized deployment.",
		HeroGIF:      "",
		Challenge:    "Needed a lightweight CMS-like system that didn't require external services. Had to support hot-reloading and zero-downtime deployments.",
		Solution:     "Built a Go-based control panel with Templ for server-side rendering, integrated CI/CD hooks, and Docker for isolated deployments.",
		Architecture: "Monolithic Go application with SQLite for data persistence, HTMX for dynamic interactions. Docker containerization enables easy deployment and scaling.",
		TechStack:    []string{"Go", "Templ", "SQLite", "HTMX", "Docker", "Make"},
		KeyFeatures: []string{
			"Real-time content editing",
			"Project metadata management",
			"Deployment automation",
			"Hot-reload capability",
			"Built-in versioning system",
		},
		LiveURL:   "https://portfolio-control.example.com",
		GitHubURL: "https://github.com/sethum-VS/portfolio-control-panel",
		Metrics: map[string]string{
			"Startup Time":    "1.2s",
			"Binary Size":     "45MB",
			"Deploy Duration": "30s",
			"Uptime":          "99.99%",
		},
	},
	{
		ID:           "mobile-banking-companion",
		Title:        "Mobile Banking Companion",
		Subtitle:     "iOS app for transaction management",
		Description:  "A native iOS application for real-time transaction tracking, budgeting, and financial insights. Emphasizes privacy with local-first architecture and end-to-end encryption.",
		HeroGIF:      "",
		Challenge:    "Balance feature richness with privacy concerns. Needed to work offline while maintaining sync capability across multiple devices.",
		Solution:     "Implemented Core Data for local persistence with CloudKit for encrypted sync. Designed reactive UI using SwiftUI with Combine for data flow management.",
		Architecture: "Native iOS architecture using MVVM pattern. SwiftUI for UI, Combine for reactive data binding, Core Data + CloudKit for storage/sync layer.",
		TechStack:    []string{"SwiftUI", "Combine", "Core Data", "CloudKit", "Firebase Analytics", "Security Framework"},
		KeyFeatures: []string{
			"Offline-first architecture",
			"End-to-end encrypted sync",
			"Real-time budget tracking",
			"Transaction categorization AI",
			"Multi-device synchronization",
		},
		LiveURL:   "https://apps.apple.com/app/banking-companion/id123456789",
		GitHubURL: "https://github.com/sethum-VS/mobile-banking-companion",
		Metrics: map[string]string{
			"App Size":     "28MB",
			"Startup Time": "0.8s",
			"Rating":       "4.8★ (500+ reviews)",
			"Active Users": "50K+",
		},
	},
}

var productsMap map[string]*Product
var mu sync.RWMutex

func init() {
	productsMap = make(map[string]*Product, len(allProducts))
	for i := range allProducts {
		productsMap[allProducts[i].ID] = &allProducts[i]
	}
}

// AllProducts returns the catalog of all portfolio projects.
func AllProducts() []Product {
	mu.RLock()
	defer mu.RUnlock()
	products := make([]Product, len(allProducts))
	copy(products, allProducts)
	return products
}

// GetProductByID finds and returns a product by its ID, or nil if not found.
func GetProductByID(id string) *Product {
	mu.RLock()
	defer mu.RUnlock()
	if p, ok := productsMap[id]; ok {
		// Return a copy to prevent accidental mutation without going through UpdateProduct
		productCopy := *p
		return &productCopy
	}
	return nil
}

// CreateProduct adds a new product to the catalog.
func CreateProduct(p Product) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := productsMap[p.ID]; exists {
		return fmt.Errorf("product with ID %s already exists", p.ID)
	}

	allProducts = append(allProducts, p)
	productsMap[p.ID] = &allProducts[len(allProducts)-1]
	return nil
}

// UpdateProduct modifies an existing product.
func UpdateProduct(id string, updated Product) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := productsMap[id]; !exists {
		return fmt.Errorf("product with ID %s not found", id)
	}

	for i, p := range allProducts {
		if p.ID == id {
			// If ID changed, update map keys
			if id != updated.ID {
				delete(productsMap, id)
			}
			allProducts[i] = updated
			productsMap[updated.ID] = &allProducts[i]
			return nil
		}
	}
	return fmt.Errorf("product with ID %s not found in list", id)
}

// DeleteProduct removes a product from the catalog.
func DeleteProduct(id string) error {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := productsMap[id]; !exists {
		return fmt.Errorf("product with ID %s not found", id)
	}

	delete(productsMap, id)
	for i, p := range allProducts {
		if p.ID == id {
			allProducts = append(allProducts[:i], allProducts[i+1:]...)
			break
		}
	}

	// Rebuild map to point to correct slice addresses since slice shifted
	productsMap = make(map[string]*Product, len(allProducts))
	for i := range allProducts {
		productsMap[allProducts[i].ID] = &allProducts[i]
	}

	return nil
}
