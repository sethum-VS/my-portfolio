package handlers

import (
	"net/http"

	"github.com/sethum-VS/my-portfolio/internal/views"
)

// HomeHandler serves the main landing page at GET /home.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request header
// is present), only the inner content fragment is returned so the browser can
// perform an outerHTML swap without re-rendering the full page shell.
// Direct browser navigation always receives the complete HTML document.
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	htmxAwareRender(w, r, views.HomeContent())
}
