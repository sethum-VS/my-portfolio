package handlers

import (
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ProductHandler serves individual product detail pages at GET /projects/{id}.
//
// Access to product pages is restricted and only intended to be accessed via
// the Projects page. Direct navigation attempts are gracefully handled.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request
// header is present), only the inner content fragment is returned.
func ProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from the route pattern /projects/{id}
	productID := r.PathValue("id")

	// Fetch product from data model
	product := models.GetProductByID(r.Context(), productID)
	if product == nil {
		http.NotFound(w, r)
		return
	}

	// Serve product page content
	htmxAwareRender(w, r, views.ProductContent(product))
}
