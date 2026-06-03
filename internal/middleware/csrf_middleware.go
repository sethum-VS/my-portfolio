package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const csrfCookieName = "csrf_token"
const csrfHeaderName = "X-CSRF-Token"

// generateCSRFToken creates a cryptographically random token.
func generateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CSRFMiddleware implements the double-submit cookie CSRF protection pattern.
// It sets a CSRF cookie on GET requests and validates the token on state-mutating
// requests (POST, PUT, DELETE) by comparing the cookie value with the X-CSRF-Token header.
//
// This is a stateless protection mechanism; no server-side token storage is required.
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For safe methods, ensure a CSRF cookie is set
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			cookie, err := r.Cookie(csrfCookieName)
			if err != nil || cookie.Value == "" {
				token, err := generateCSRFToken()
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:     csrfCookieName,
					Value:    token,
					Path:     "/",
					HttpOnly: false, // Must be readable by JS to send in header
					Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
					SameSite: http.SameSiteStrictMode,
				})
			}
			next.ServeHTTP(w, r)
			return
		}

		// For state-mutating methods, validate the CSRF token
		cookie, err := r.Cookie(csrfCookieName)
		if err != nil || cookie.Value == "" {
			csrfForbidden(w, r, "Session expired. Refresh the page and try again.")
			return
		}

		headerToken := r.Header.Get(csrfHeaderName)
		if headerToken == "" {
			csrfForbidden(w, r, "Security token missing. Refresh the page and try again.")
			return
		}

		if cookie.Value != headerToken {
			csrfForbidden(w, r, "Security token mismatch. Refresh the page and try again.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func csrfForbidden(w http.ResponseWriter, r *http.Request, message string) {
	if r.Header.Get("HX-Request") == "true" && r.URL.Path == "/api/resume/request" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`<div id="resume-modal-inner" class="relative w-full max-w-md rounded-2xl border border-white/10 bg-zinc-900/40 p-8 shadow-2xl backdrop-blur-xl"><p class="text-sm text-red-400 font-body rounded-lg border border-red-500/30 bg-red-500/10 px-3 py-2" role="alert">` + message + `</p></div>`))
		return
	}
	http.Error(w, "Forbidden: "+message, http.StatusForbidden)
}
