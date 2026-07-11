package handlers

import (
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ResumeModalHandler serves GET /modal/resume.
func ResumeModalHandler(w http.ResponseWriter, r *http.Request) {
	cfg := models.GetResumeConfig()
	siteKey := os.Getenv("TURNSTILE_SITE_KEY")
	templ.Handler(views.ResumeModal(cfg.IsComingSoon, siteKey, "")).ServeHTTP(w, r)
}
