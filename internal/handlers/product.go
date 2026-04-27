package handlers

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ProductHandler serves individual product detail pages at GET /projects/:id.
//
// Access to product pages is restricted and only intended to be accessed via
// the Projects page. Direct navigation attempts are gracefully handled.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request
// header is present), only the inner content fragment is returned.
func ProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	// Expected format: /projects/product-id
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/projects/"), "/")
	productID := pathParts[0]

	// Fetch product from data model
	product := models.GetProductByID(productID)
	if product == nil {
		http.NotFound(w, r)
		return
	}

	// Serve product page content
	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(views.ProductContent(product)).ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
