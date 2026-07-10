package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sethum-VS/my-portfolio/internal/services"
)

// AdminAuthMiddleware protects administrative routes by verifying a Supabase session token
// and checking the user's email against an authorized whitelist loaded from environment variables.
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read the session_token cookie
		cookie, err := r.Cookie("session_token")
		if err != nil {
			log.Printf("Auth Error: No session cookie found: %v", err)
			// If it's an HTMX request, we use the HX-Redirect header
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Verify the session cookie using Supabase JWT verification
		token, err := services.VerifySupabaseJWT(cookie.Value)
		if err != nil || !token.Valid {
			log.Printf("Auth Error: Supabase cookie verification failed: %v", err)
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:   "session_token",
				Value:  "",
				Path:   "/",
				MaxAge: -1,
			})
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("Auth Error: Invalid JWT claims format")
			http.Error(w, "Forbidden: Invalid token", http.StatusForbidden)
			return
		}

		// Whitelist verification from environment variable
		adminEmailsRaw := os.Getenv("ADMIN_EMAIL")
		if adminEmailsRaw == "" {
			log.Println("CRITICAL SECURITY WARNING: ADMIN_EMAIL environment variable is not set. Access denied to all users.")
			http.Error(w, "Forbidden: Administrative access is currently disabled for security.", http.StatusForbidden)
			return
		}

		// Parse multi-admin support (comma-separated)
		authorized := false
		email, ok := claims["email"].(string)
		if ok {
			adminEmails := strings.Split(adminEmailsRaw, ",")
			for _, adminEmail := range adminEmails {
				if email == strings.TrimSpace(adminEmail) {
					authorized = true
					break
				}
			}
		}

		if !authorized {
			log.Printf("Auth Error: Unauthorized access attempt by %s", email)
			// Clear invalid cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login?error=unauthorized")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/login?error=unauthorized", http.StatusSeeOther)
			return
		}

		// If everything is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
