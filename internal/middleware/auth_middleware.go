package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sethum-VS/my-portfolio/internal/services"
)

// AdminAuthMiddleware protects administrative routes by verifying a Firebase session cookie
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

		// Verify the token using Firebase Admin
		token, err := services.FirebaseAuth.VerifyIDToken(r.Context(), cookie.Value)
		if err != nil {
			log.Printf("Auth Error: Token verification failed: %v", err)
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

		// Whitelist verification from environment variable
		adminEmailsRaw := os.Getenv("ADMIN_EMAIL")
		if adminEmailsRaw == "" {
			log.Println("CRITICAL SECURITY WARNING: ADMIN_EMAIL environment variable is not set. Access denied to all users.")
			http.Error(w, "Forbidden: Administrative access is currently disabled for security.", http.StatusForbidden)
			return
		}

		// Parse multi-admin support (comma-separated)
		authorized := false
		email, ok := token.Claims["email"].(string)
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
			http.Error(w, "Forbidden: Administrative access required", http.StatusForbidden)
			return
		}

		// If everything is valid, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
