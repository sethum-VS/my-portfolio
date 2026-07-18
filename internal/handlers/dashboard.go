package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// DashboardHandler serves the protected administrative dashboard.
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	products := models.AllProducts(r.Context())
	templ.Handler(views.DashboardPage(products)).ServeHTTP(w, r)
}

// HandleAIParseReadme processes raw README content and returns the auto-filled form fields.
func HandleAIParseReadme(w http.ResponseWriter, r *http.Request) {

	readme := r.FormValue("readme_content")
	if readme == "" {
		readme = r.FormValue("description")
	}

	if readme == "" {
		log.Println("AI Parse Error: No README content provided in form values")
		http.Error(w, "No README content provided", http.StatusBadRequest)
		return
	}

	log.Printf("AI Parse: Starting README analysis (Length: %d characters)...", len(readme))
	product, err := services.ParseReadmeToProductContext(r.Context(), readme)
	if err != nil {
		log.Printf("AI Parse Error: %v", err)
		http.Error(w, fmt.Sprintf("AI parsing failed: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("AI Parse Success: Extracted context for '%s'", product.Title)

	// Return only the form fields component
	templ.Handler(views.DashboardFormFields(product)).ServeHTTP(w, r)
}
