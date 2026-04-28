package handlers

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// DashboardHandler serves the protected administrative dashboard.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	products := models.AllProducts()
	templ.Handler(views.DashboardPage(products)).ServeHTTP(w, r)
}
