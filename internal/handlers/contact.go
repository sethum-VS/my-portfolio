package handlers

import (
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ContactHandler serves the Contact page at GET /contact.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request header
// is present), only the inner content fragment is returned.
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	htmxAwareRender(w, r, views.ContactContent())
}
