package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// DashboardHandler is a protected route placeholder for future sprints.
// In Sprint 2+ this will be guarded by a session/JWT middleware.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(views.DashboardPage()).ServeHTTP(w, r)
}
