package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// SupabaseConfig holds the public config needed for the JS client
type SupabaseConfig struct {
	URL     string `json:"url"`
	AnonKey string `json:"anon_key"`
}

// LoginHandler serves the login page with Supabase config injected from env vars
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseAnonKey := os.Getenv("SUPABASE_ANON_KEY")
	templ.Handler(views.LoginPage(supabaseURL, supabaseAnonKey)).ServeHTTP(w, r)
}

// HandleCreateSession processes a Supabase access token (JWT) and creates an opaque
// server-side session cookie. 
func HandleCreateSession(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		IDToken string `json:"idToken"` // It's actually the Supabase access token now
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if payload.IDToken == "" {
		http.Error(w, "idToken is required", http.StatusBadRequest)
		return
	}

	// Just set the cookie with the access token. 
	// The middleware handles verification.
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    payload.IDToken,
		MaxAge:   3600, // 1 hour (Supabase tokens usually expire in 1h anyway)
		HttpOnly: true,
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
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	
	// Redirect to home or login after logout
	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
