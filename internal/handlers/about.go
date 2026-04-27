package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// AboutHandler serves the About page at GET /about.
//
// HTMX-aware: when the request originates from an HTMX swap (HX-Request header
// is present), only the inner content fragment is returned.
func AboutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("HX-Request") == "true" {
		templ.Handler(views.AboutContent()).ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
