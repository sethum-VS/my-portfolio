package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// AdminPlaceholderHandler serves the dashboard standby placeholder.
func AdminPlaceholderHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(views.DashboardPlaceholder()).ServeHTTP(w, r)
}

// AdminProjectFormHandler handles GET /api/projects/{id} to load the form.
func AdminProjectFormHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" || pathParts[0] == "new" {
		// Render empty form
		templ.Handler(views.DashboardProjectForm(nil)).ServeHTTP(w, r)
		return
	}
	
	id := pathParts[0]
	product := models.GetProductByID(id)
	if product == nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}
	
	templ.Handler(views.DashboardProjectForm(product)).ServeHTTP(w, r)
}

// AdminProjectDeleteHandler handles DELETE /api/projects/{id}.
func AdminProjectDeleteHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}
	
	id := pathParts[0]
	err := models.DeleteProduct(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Re-render the project list
	products := models.AllProducts()
	templ.Handler(views.DashboardProjectList(products, "")).ServeHTTP(w, r)
}

// AdminProjectSaveHandler handles POST /api/projects to create or update.
func AdminProjectSaveHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	
	// Create Product model from form
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "Project ID is required", http.StatusBadRequest)
		return
	}

	isUpdate := r.FormValue("is_update") == "true"
	originalID := r.FormValue("original_id")

	p := models.Product{
		ID:           id,
		Title:        r.FormValue("title"),
		Subtitle:     r.FormValue("subtitle"),
		Challenge:    r.FormValue("challenge"),
		Solution:     r.FormValue("solution"),
		Architecture: r.FormValue("architecture"),
		TechStack:    parseCommaSeparated(r.FormValue("tech_stack")),
		KeyFeatures:  parseCommaSeparated(r.FormValue("core_features")),
		HeroGIF:      r.FormValue("hero_gif"),
		Description:  r.FormValue("description"),
	}

	if isUpdate && originalID != "" {
		err = models.UpdateProduct(originalID, p)
	} else {
		err = models.CreateProduct(p)
	}
	
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "<div id=\"form-response\" class=\"text-error\">Error: %s</div>", err.Error())
		return
	}
	
	// We want to update the form (to show success/updated state) AND the list (OOB)
	products := models.AllProducts()
	
	// Setup OOB response for the list
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte("<div id=\"dashboard-project-list\" hx-swap-oob=\"true\">"))
	views.DashboardProjectList(products, id).Render(r.Context(), w)
	w.Write([]byte("</div>"))
	
	// Normal response for the form
	views.DashboardProjectForm(&p).Render(r.Context(), w)
}

func parseCommaSeparated(val string) []string {
	if strings.TrimSpace(val) == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	var res []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			res = append(res, trimmed)
		}
	}
	return res
}
