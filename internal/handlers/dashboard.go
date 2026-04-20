package handlers

import (
	"fmt"
	"net/http"
)

// DashboardHandler is a protected route placeholder for future sprints.
// In Sprint 2+ this will be guarded by a session/JWT middleware.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Dashboard — Protected Route (Sprint 2 Placeholder)")
}
