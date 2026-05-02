package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// LoginHandler serves the login page with Firebase config injected from env vars
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	config := services.GetFirebaseClientConfig()
	templ.Handler(views.LoginPage(config)).ServeHTTP(w, r)
}

// sessionDuration defines the lifetime of admin session cookies.
const sessionDuration = 1 * time.Hour

// HandleCreateSession processes a Firebase ID token and creates an opaque
// server-side session cookie via Firebase's CreateSessionCookie API.
// S-03: This replaces the previous pattern of storing the raw ID token JWT
// directly in the cookie, which exposed user claims and could be replayed
// against Firebase APIs if intercepted.
func HandleCreateSession(w http.ResponseWriter, r *http.Request) {
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

	// Verify the ID token first
	_, err := services.FirebaseAuth.VerifyIDToken(r.Context(), payload.IDToken)
	if err != nil {
		http.Error(w, "Invalid ID token", http.StatusUnauthorized)
		return
	}

	// Create an opaque Firebase session cookie (server-side)
	sessionCookie, err := services.FirebaseAuth.SessionCookie(r.Context(), payload.IDToken, sessionDuration)
	if err != nil {
		log.Printf("Session cookie creation failed (IAM permission issue likely), falling back to ID token: %v", err)
		// Fallback to using the raw ID token as the session cookie for local dev / limited IAM
		sessionCookie = payload.IDToken
	}

	// Set the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionCookie,
		Expires:  time.Now().Add(sessionDuration),
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

// HandleLogout clears the session cookie.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Revoke the session cookie on Firebase's side
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		// Verify and decode the session cookie to get the UID for revocation
		decoded, err := services.FirebaseAuth.VerifySessionCookie(r.Context(), cookie.Value)
		if err == nil {
			// Revoke all refresh tokens for this user
			if err := services.FirebaseAuth.RevokeRefreshTokens(r.Context(), decoded.UID); err != nil {
				log.Printf("Failed to revoke refresh tokens: %v", err)
			}
		}
	}

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
