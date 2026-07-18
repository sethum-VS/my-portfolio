package handlers

import (
	"context"
	"fmt"
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

func renderResumeAdmin(w http.ResponseWriter, r *http.Request, cfg models.ResumeConfig, message string, isError bool) {
	templ.Handler(views.ResumeAdminCard(cfg, models.WaitlistCount(r.Context()), message, isError)).ServeHTTP(w, r)
}

// AdminResumeFormHandler serves GET /api/admin/resume.
func AdminResumeFormHandler(w http.ResponseWriter, r *http.Request) {
	cfg := models.GetResumeConfig(r.Context())
	renderResumeAdmin(w, r, cfg, "", false)
}

// AdminResumeBroadcastHandler sends the current PDF to everyone on the waitlist.
func AdminResumeBroadcastHandler(w http.ResponseWriter, r *http.Request) {
	cfg := models.GetResumeConfig(r.Context())
	waitlistCount := models.WaitlistCount(r.Context())

	if waitlistCount == 0 {
		renderResumeAdmin(w, r, cfg, "No one is on the waitlist.", true)
		return
	}
	if cfg.PDFStorageURI == "" {
		renderResumeAdmin(w, r, cfg, "Upload a PDF before sending to the waitlist.", true)
		return
	}

	pdfURI := cfg.PDFStorageURI
	go services.BroadcastResumeToWaitlist(context.Background(), pdfURI)

	msg := fmt.Sprintf("Sending CV to %d waitlist subscriber(s) in the background.", waitlistCount)
	renderResumeAdmin(w, r, cfg, msg, false)
}

// AdminResumeSaveHandler handles POST /api/admin/resume.
func AdminResumeSaveHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxResumeUploadBytes); err != nil {
		renderResumeAdmin(w, r, models.GetResumeConfig(r.Context()), "Invalid form data.", true)
		return
	}

	prev := models.GetResumeConfig(r.Context())
	cfg := prev
	cfg.IsComingSoon = r.FormValue("is_coming_soon") == "on"

	file, header, err := r.FormFile("resume_pdf")
	if err == nil {
		defer file.Close()

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".pdf" {
			renderResumeAdmin(w, r, prev, "Only PDF files are allowed.", true)
			return
		}

		contentType := header.Header.Get("Content-Type")
		if contentType != "" && contentType != "application/pdf" {
			renderResumeAdmin(w, r, prev, "Invalid file type. Upload a PDF.", true)
			return
		}

		gsURI, err := services.UploadResumePDF(r.Context(), file, "application/pdf")
		if err != nil {
			log.Printf("resume upload error: %v", err)
			renderResumeAdmin(w, r, prev, "Failed to upload PDF. Check GCS configuration.", true)
			return
		}
		cfg.PDFStorageURI = gsURI
	}

	if err := models.SaveResumeConfig(r.Context(), cfg); err != nil {
		log.Printf("resume config save error: %v", err)
		renderResumeAdmin(w, r, prev, "Failed to save configuration.", true)
		return
	}

	message := "Resume settings saved."
	waitlistCount := models.WaitlistCount(r.Context())
	turnedOffComingSoon := prev.IsComingSoon && !cfg.IsComingSoon
	manualBroadcast := r.FormValue("send_to_waitlist") == "on"

	if cfg.PDFStorageURI != "" && waitlistCount > 0 && (turnedOffComingSoon || manualBroadcast) {
		pdfURI := cfg.PDFStorageURI
		go services.BroadcastResumeToWaitlist(context.Background(), pdfURI)
		if turnedOffComingSoon {
			message = fmt.Sprintf("Resume settings saved. Sending CV to %d waitlist subscriber(s).", waitlistCount)
		} else {
			message = fmt.Sprintf("Resume settings saved. Sending CV to %d waitlist subscriber(s) in the background.", waitlistCount)
		}
	}

	renderResumeAdmin(w, r, cfg, message, false)
}
