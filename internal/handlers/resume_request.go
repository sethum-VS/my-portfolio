package handlers

import (
	"log"
	"net/http"
	"net/mail"
	"os"
	"strings"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/models"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// ResumeRequestHandler handles POST /api/resume/request.
func ResumeRequestHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		resumeRequestError(w, r, "Invalid form data.", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	token := r.FormValue("g-recaptcha-response")

	if email == "" {
		resumeRequestError(w, r, "Email is required.", http.StatusBadRequest)
		return
	}
	if _, err := mail.ParseAddress(email); err != nil {
		resumeRequestError(w, r, "Invalid email address.", http.StatusBadRequest)
		return
	}

	if _, err := services.VerifyRecaptcha(r.Context(), token, r); err != nil {
		log.Printf("recaptcha verification failed: %v", err)
		resumeRequestError(w, r, "Verification failed. Complete reCAPTCHA and try again.", http.StatusForbidden)
		return
	}

	cfg := models.GetResumeConfig()

	if cfg.IsComingSoon {
		if err := models.AddToWaitlist(email); err != nil {
			log.Printf("waitlist error: %v", err)
			resumeRequestError(w, r, "Could not add to waitlist. Try again later.", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		templ.Handler(views.ResumeModalSuccess("Added to waitlist.")).ServeHTTP(w, r)
		return
	}

	if cfg.PDFStorageURI == "" {
		resumeRequestError(w, r, "Resume is not available yet.", http.StatusServiceUnavailable)
		return
	}

	pdfBytes, err := services.DownloadResumePDF(r.Context(), cfg.PDFStorageURI)
	if err != nil {
		log.Printf("resume download error: %v", err)
		resumeRequestError(w, r, "Could not retrieve resume. Try again later.", http.StatusInternalServerError)
		return
	}

	if err := services.SendEmail(r.Context(), email, services.ResumeEmailSubject(), services.ResumeEmailHTML(), pdfBytes, "Seth_Ummethsanda_CV.pdf"); err != nil {
		log.Printf("resume email error: %v", err)
		resumeRequestError(w, r, "Could not send email. Try again later.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	templ.Handler(views.ResumeModalSuccess("Check your inbox — your CV is on the way.")).ServeHTTP(w, r)
}

func resumeRequestError(w http.ResponseWriter, r *http.Request, message string, status int) {
	if r.Header.Get("HX-Request") != "true" {
		http.Error(w, message, status)
		return
	}

	cfg := models.GetResumeConfig()
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	w.WriteHeader(status)
	templ.Handler(views.ResumeModalInner(cfg.IsComingSoon, siteKey, message)).ServeHTTP(w, r)
}
