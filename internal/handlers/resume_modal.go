package handlers

import (
	"github.com/sethum-VS/my-portfolio/internal/config"

	"net/http"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ResumeModalHandler serves GET /modal/resume.
func ResumeModalHandler(w http.ResponseWriter, r *http.Request) {
	cfg := models.GetResumeConfig(r.Context())
	siteKey := config.AppConfig.TurnstileSiteKey
	templ.Handler(views.ResumeModal(cfg.IsComingSoon, siteKey, "")).ServeHTTP(w, r)
}
