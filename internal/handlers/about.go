package handlers

import (
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/views"
)

// AboutHandler serves the About page at GET /about.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request header
// is present), only the inner content fragment is returned.
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	htmxAwareRender(w, r, views.AboutContent())
}
