package handlers

import (
	"net/http"

	"github.com/a-h/templ"
)

// htmxAwareRender serves the given templ component if the request comes from an
// HTMX swap (HX-Request header), otherwise redirects to the splash page.
// This eliminates the repeated if/else pattern across all public handlers.
func htmxAwareRender(w http.ResponseWriter, r *http.Request, content templ.Component) {
	if r.Header.Get("HX-Request") == "true" {
		content.Render(r.Context(), w)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}
