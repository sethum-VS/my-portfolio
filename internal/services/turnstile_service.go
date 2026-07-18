package services

import (
	"github.com/sethum-VS/my-portfolio/internal/config"

	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// TurnstileResponse represents the verification response from Cloudflare Turnstile API.
type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
	Action      string   `json:"action"`
}

var turnstileVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

func requestClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if i := strings.IndexByte(forwarded, ','); i >= 0 {
			return strings.TrimSpace(forwarded[:i])
		}
		return strings.TrimSpace(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// VerifyTurnstile validates a Cloudflare Turnstile token.
func VerifyTurnstile(ctx context.Context, token string, r *http.Request) error {
	if token == "" {
		return fmt.Errorf("missing turnstile token")
	}

	secretKey := config.AppConfig.TurnstileSecretKey
	if secretKey == "" {
		return fmt.Errorf("turnstile secret key is not configured")
	}

	apiURL := turnstileVerifyURL

	data := url.Values{}
	data.Set("secret", secretKey)
	data.Set("response", token)

	if remoteIP := requestClientIP(r); remoteIP != "" {
		data.Set("remoteip", remoteIP)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("turnstile request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("turnstile API returned status code: %d", resp.StatusCode)
	}

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode turnstile response: %w", err)
	}

	if !result.Success {
		return fmt.Errorf("turnstile verification failed: %s", strings.Join(result.ErrorCodes, ", "))
	}

	return nil
}
