package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sethum-VS/my-portfolio/internal/config"
)

func TestVerifyTurnstile(t *testing.T) {
	// Set mock secret key
	os.Setenv("TURNSTILE_SECRET_KEY", "mock-secret-key")
	config.Load()
	defer func() {
		os.Unsetenv("TURNSTILE_SECRET_KEY")
		config.Load()
	}()

	t.Run("missing token", func(t *testing.T) {
		err := VerifyTurnstile(context.Background(), "", nil)
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if err.Error() != "missing turnstile token" {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("successful verification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST request, got %s", r.Method)
			}
			if err := r.ParseForm(); err != nil {
				t.Errorf("failed to parse form: %v", err)
			}
			if r.FormValue("secret") != "mock-secret-key" {
				t.Errorf("unexpected secret: %s", r.FormValue("secret"))
			}
			if r.FormValue("response") != "valid-token" {
				t.Errorf("unexpected response token: %s", r.FormValue("response"))
			}
			if r.FormValue("remoteip") != "127.0.0.1" {
				t.Errorf("unexpected remoteip: %s", r.FormValue("remoteip"))
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TurnstileResponse{
				Success: true,
			})
		}))
		defer server.Close()

		// Temporarily point to mock server
		origURL := turnstileVerifyURL
		turnstileVerifyURL = server.URL
		defer func() { turnstileVerifyURL = origURL }()

		req := httptest.NewRequest("POST", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"

		err := VerifyTurnstile(context.Background(), "valid-token", req)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("failed verification", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(TurnstileResponse{
				Success:    false,
				ErrorCodes: []string{"invalid-input-response"},
			})
		}))
		defer server.Close()

		// Temporarily point to mock server
		origURL := turnstileVerifyURL
		turnstileVerifyURL = server.URL
		defer func() { turnstileVerifyURL = origURL }()

		err := VerifyTurnstile(context.Background(), "invalid-token", nil)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "turnstile verification failed: invalid-input-response" {
			t.Errorf("unexpected error: %v", err)
		}
	})
}
