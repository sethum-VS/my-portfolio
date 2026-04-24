package main

import (
	"log"
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	// ── Static file server ──────────────────────────────────────────────────
	// Serves everything under /static/ from the ./static directory on disk.
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// ── Application routes ──────────────────────────────────────────────────
	// Go 1.22+ enhanced pattern syntax: "METHOD /path"
	mux.HandleFunc("GET /", handlers.SplashHandler)
	mux.HandleFunc("GET /home", handlers.HomeHandler)
	mux.HandleFunc("GET /about", handlers.AboutHandler)
	mux.HandleFunc("GET /dashboard", handlers.DashboardHandler)

	const addr = ":8080"
	log.Printf("→ Server listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
