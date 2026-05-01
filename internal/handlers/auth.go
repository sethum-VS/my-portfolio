package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// LoginHandler serves the login page
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	templ.Handler(views.LoginPage()).ServeHTTP(w, r)
}

// HandleCreateSession processes a Firebase ID token and sets a secure session cookie.
func HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		IDToken string `json:"idToken"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.IDToken == "" {
		http.Error(w, "idToken is required", http.StatusBadRequest)
		return
	}

	// Verify the ID token using Firebase Admin SDK
	_, err := services.FirebaseAuth.VerifyIDToken(r.Context(), payload.IDToken)
	if err != nil {
		http.Error(w, "Invalid ID token", http.StatusUnauthorized)
		return
	}

	// Set a secure, HttpOnly session cookie
	// We use the ID token as the session token for simplicity in this monolith.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    payload.IDToken, // Store the token string
		Expires:  time.Now().Add(1 * time.Hour),
		HttpOnly: true,
		// Secure should be true in production (HTTPS), but false for local HTTP testing
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// HandleLogout clears the session cookie.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	
	// Redirect to home or login after logout
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
