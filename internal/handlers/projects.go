package handlers

import (
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ProjectsHandler serves the Projects page at GET /projects.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request
// header is present), only the inner content fragment is returned.
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	products := models.AllProducts(r.Context())
	htmxAwareRender(w, r, views.ProjectsContent(products))
}
