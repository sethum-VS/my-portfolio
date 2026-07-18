package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/a-h/templ"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sethum-VS/my-portfolio/internal/config"
	"github.com/sethum-VS/my-portfolio/internal/services"
	"github.com/sethum-VS/my-portfolio/internal/views"
)

// SupabaseConfig holds the public config needed for the JS client
type SupabaseConfig struct {
	URL     string `json:"url"`
	AnonKey string `json:"anon_key"`
}

// LoginHandler serves the login page with Supabase config injected from config
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	supabaseURL := config.AppConfig.SupabaseURL
	supabaseAnonKey := config.AppConfig.SupabaseAnonKey
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

	// Verify the token and check authorization before creating the session
	token, err := services.VerifySupabaseJWT(payload.IDToken)
	if err != nil || !token.Valid {
		log.Printf("Auth Error: Supabase token verification failed during session creation: %v", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Whitelist verification from centralized config
	if len(config.AppConfig.AdminEmails) == 0 {
		log.Println("CRITICAL SECURITY WARNING: ADMIN_EMAIL is not set in config. Access denied to all users.")
		http.Error(w, "Forbidden: Administrative access is currently disabled for security.", http.StatusForbidden)
		return
	}

	authorized := false
	email, ok := claims["email"].(string)
	if ok {
		for _, adminEmail := range config.AppConfig.AdminEmails {
			if email == adminEmail {
				authorized = true
				break
			}
		}
	}

	if !authorized {
		log.Printf("Auth Error: Unauthorized session creation attempt by %s", email)
		http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
		return
	}

	// Set the cookie with the access token.
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
