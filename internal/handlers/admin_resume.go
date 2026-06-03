package handlers

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

const maxResumeUploadBytes = 10 << 20 // 10MB

// AdminResumeFormHandler serves GET /api/admin/resume.
func AdminResumeFormHandler(w http.ResponseWriter, r *http.Request) {
	cfg := models.GetResumeConfig()
	templ.Handler(views.ResumeAdminCard(cfg, "", false)).ServeHTTP(w, r)
}

// AdminResumeSaveHandler handles POST /api/admin/resume.
func AdminResumeSaveHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxResumeUploadBytes); err != nil {
		templ.Handler(views.ResumeAdminCard(models.GetResumeConfig(), "Invalid form data.", true)).ServeHTTP(w, r)
		return
	}

	prev := models.GetResumeConfig()
	cfg := prev
	cfg.IsComingSoon = r.FormValue("is_coming_soon") == "on"

	file, header, err := r.FormFile("resume_pdf")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".pdf" {
			templ.Handler(views.ResumeAdminCard(prev, "Only PDF files are allowed.", true)).ServeHTTP(w, r)
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType != "" && contentType != "application/pdf" {
			templ.Handler(views.ResumeAdminCard(prev, "Invalid file type. Upload a PDF.", true)).ServeHTTP(w, r)
			return
		}

		gsURI, err := services.UploadResumePDF(r.Context(), file, "application/pdf")
		if err != nil {
			log.Printf("resume upload error: %v", err)
			templ.Handler(views.ResumeAdminCard(prev, "Failed to upload PDF. Check GCS configuration.", true)).ServeHTTP(w, r)
			return
		}
		cfg.PDFStorageURI = gsURI
	}

	if err := models.SaveResumeConfig(cfg); err != nil {
		log.Printf("resume config save error: %v", err)
		templ.Handler(views.ResumeAdminCard(prev, "Failed to save configuration.", true)).ServeHTTP(w, r)
		return
	}

	if prev.IsComingSoon && !cfg.IsComingSoon && cfg.PDFStorageURI != "" {
		pdfURI := cfg.PDFStorageURI
		go services.BroadcastResumeToWaitlist(context.Background(), pdfURI)
	}

	templ.Handler(views.ResumeAdminCard(cfg, "Resume settings saved.", false)).ServeHTTP(w, r)
}
