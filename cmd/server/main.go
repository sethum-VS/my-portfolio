package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/joho/godotenv"

	"github.com/sethum-VS/my-portfolio/internal/handlers"
	"github.com/sethum-VS/my-portfolio/internal/middleware"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/services"
)

// securityHeadersMiddleware adds the missing security headers.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com https://unpkg.com https://cdn.jsdelivr.net https://www.gstatic.com https://apis.google.com blob:; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: blob: https://lh3.googleusercontent.com https://github.com https://raw.githubusercontent.com; connect-src 'self' https://cdn.jsdelivr.net https://*.googleapis.com https://www.gstatic.com https://*.firebaseio.com https://*.firebaseapp.com; frame-src https://*.firebaseapp.com https://apis.google.com; worker-src 'self' blob:;")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// cacheControlMiddleware adds caching headers for static assets.
func cacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Immutable assets (CSS/JS) cache for 1 year
		if strings.HasSuffix(r.URL.Path, ".js") || strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		} else if strings.HasPrefix(r.URL.Path, "/static/images/") {
			w.Header().Set("Cache-Control", "public, max-age=86400") // 1 day for images
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or could not be loaded: %v", err)
	}

	// Initialize Firebase & Firestore
	if err := services.InitFirebase(context.Background()); err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	
	// Inject Firestore client into the models package
	models.InitDB(services.FirestoreClient)

	mux := http.NewServeMux()

	// ── Static file server ──────────────────────────────────────────────────
	// Serves everything under /static/ from the ./static directory on disk.
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.Handle("GET /static/", cacheControlMiddleware(staticHandler))

	// ── Public Application routes ────────────────────────────────────────────
	mux.HandleFunc("GET /", handlers.SplashHandler)
	mux.HandleFunc("GET /home", handlers.HomeHandler)
	mux.HandleFunc("GET /about", handlers.AboutHandler)
	mux.HandleFunc("GET /projects", handlers.ProjectsHandler)
	mux.HandleFunc("GET /projects/{id}", handlers.ProductHandler)
	mux.HandleFunc("GET /contact", handlers.ContactHandler)
	
	// Authentication
	mux.HandleFunc("GET /login", handlers.LoginHandler)
	mux.HandleFunc("POST /api/auth/session", handlers.HandleCreateSession)
	mux.HandleFunc("POST /api/auth/logout", handlers.HandleLogout)

	// ── Protected Admin Dashboard API ────────────────────────────────────────
	// We wrap these handlers in the AdminAuthMiddleware
	mux.Handle("GET /dashboard", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.DashboardHandler)))
	mux.Handle("GET /api/dashboard/placeholder", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.AdminPlaceholderHandler)))
	mux.Handle("GET /api/projects/{id}", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.AdminProjectFormHandler)))
	mux.Handle("POST /api/projects", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.AdminProjectSaveHandler)))
	mux.Handle("DELETE /api/projects/{id}", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.AdminProjectDeleteHandler)))
	mux.Handle("POST /api/ai/parse-readme", middleware.AdminAuthMiddleware(http.HandlerFunc(handlers.HandleAIParseReadme)))

	secureMux := securityHeadersMiddleware(mux)

	const addr = ":8080"
	log.Printf("→ Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, secureMux))
}
