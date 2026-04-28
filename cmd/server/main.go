package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/sethum-VS/my-portfolio/internal/handlers"
)

// securityHeadersMiddleware adds the missing security headers.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: blob: https://lh3.googleusercontent.com; connect-src 'self'")
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
	mux := http.NewServeMux()

	// ── Static file server ──────────────────────────────────────────────────
	// Serves everything under /static/ from the ./static directory on disk.
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))
	mux.Handle("GET /static/", cacheControlMiddleware(staticHandler))

	// ── Application routes ──────────────────────────────────────────────────
	// Go 1.22+ enhanced pattern syntax: "METHOD /path"
	mux.HandleFunc("GET /", handlers.SplashHandler)
	mux.HandleFunc("GET /home", handlers.HomeHandler)
	mux.HandleFunc("GET /about", handlers.AboutHandler)
	mux.HandleFunc("GET /projects", handlers.ProjectsHandler)
	mux.HandleFunc("GET /projects/{id}", handlers.ProductHandler)
	mux.HandleFunc("GET /contact", handlers.ContactHandler)
	mux.HandleFunc("GET /dashboard", handlers.DashboardHandler)

	secureMux := securityHeadersMiddleware(mux)

	const addr = ":8080"
	log.Printf("→ Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, secureMux))
}
