package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// SplashHandler serves the animated splash screen at GET /.
// Because net/http's "/" pattern is a catch-all in Go's ServeMux, we
// explicitly 404 any path that isn't exactly "/".
func SplashHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	templ.Handler(views.SplashPage()).ServeHTTP(w, r)
}
