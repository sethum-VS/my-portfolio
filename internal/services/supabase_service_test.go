package services

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sethum-VS/my-portfolio/internal/config"
)

func TestVerifySupabaseJWT_HS256(t *testing.T) {
	// Initialize the secret before running VerifySupabaseJWT
	secret := "test-jwt-secret-key-with-sufficient-length"
	os.Setenv("SUPABASE_JWT_SECRET", secret)
	config.Load()
	defer func() {
		os.Unsetenv("SUPABASE_JWT_SECRET")
		config.Load()
	}()

	// Helper to generate a token
	createToken := func(claims jwt.MapClaims, signingKey []byte) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		return token.SignedString(signingKey)
	}

	t.Run("valid token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tokenStr, err := createToken(claims, []byte(secret))
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		token, err := VerifySupabaseJWT(tokenStr)
		if err != nil {
			t.Fatalf("expected valid token, got error: %v", err)
		}

		if !token.Valid {
			t.Fatal("expected token to be valid")
		}

		parsedClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("expected map claims")
		}

		if parsedClaims["sub"] != "user-123" {
			t.Errorf("expected sub 'user-123', got '%v'", parsedClaims["sub"])
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tokenStr, err := createToken(claims, []byte("wrong-secret"))
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		_, err = VerifySupabaseJWT(tokenStr)
		if err == nil {
			t.Fatal("expected error for invalid signature, got nil")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		claims := jwt.MapClaims{
			"sub": "user-123",
			"exp": time.Now().Add(-time.Hour).Unix(),
		}
		tokenStr, err := createToken(claims, []byte(secret))
		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		_, err = VerifySupabaseJWT(tokenStr)
		if err == nil {
			t.Fatal("expected error for expired token, got nil")
		}
	})
}

func TestVerifySupabaseJWT_ES256(t *testing.T) {
	// 1. Generate ES256 private key for signing
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ECDSA key: %v", err)
	}

	kid := "test-key-id-123"

	// 2. Start mock JWKS server
	jwksHandler := func(w http.ResponseWriter, r *http.Request) {
		pubKey := privKey.Public().(*ecdsa.PublicKey)
		jwk := JWK{
			Alg: "ES256",
			Crv: "P-256",
			Kid: kid,
			Kty: "EC",
			X:   base64.RawURLEncoding.EncodeToString(pubKey.X.Bytes()),
			Y:   base64.RawURLEncoding.EncodeToString(pubKey.Y.Bytes()),
		}
		jwks := JWKS{
			Keys: []JWK{jwk},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	}
	server := httptest.NewServer(http.HandlerFunc(jwksHandler))
	defer server.Close()

	// Configure environment variables
	os.Setenv("SUPABASE_URL", server.URL)
	config.Load()
	defer func() {
		os.Unsetenv("SUPABASE_URL")
		config.Load()
	}()

	// Helper to generate an ES256 token
	createES256Token := func(claims jwt.MapClaims, key *ecdsa.PrivateKey, kid string) (string, error) {
		token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
		token.Header["kid"] = kid
		return token.SignedString(key)
	}

	t.Run("valid ES256 token", func(t *testing.T) {
		// Reset jwkCache to force a fresh fetch
		jwkCacheMu.Lock()
		jwkCache = nil
		jwkCacheMu.Unlock()

		claims := jwt.MapClaims{
			"sub": "user-456",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tokenStr, err := createES256Token(claims, privKey, kid)
		if err != nil {
			t.Fatalf("failed to create ES256 token: %v", err)
		}

		token, err := VerifySupabaseJWT(tokenStr)
		if err != nil {
			t.Fatalf("expected valid ES256 token, got error: %v", err)
		}

		if !token.Valid {
			t.Fatal("expected token to be valid")
		}

		parsedClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			t.Fatal("expected map claims")
		}

		if parsedClaims["sub"] != "user-456" {
			t.Errorf("expected sub 'user-456', got '%v'", parsedClaims["sub"])
		}
	})

	t.Run("invalid kid", func(t *testing.T) {
		// Reset jwkCache
		jwkCacheMu.Lock()
		jwkCache = nil
		jwkCacheMu.Unlock()

		claims := jwt.MapClaims{
			"sub": "user-456",
			"exp": time.Now().Add(time.Hour).Unix(),
		}
		tokenStr, err := createES256Token(claims, privKey, "unknown-kid")
		if err != nil {
			t.Fatalf("failed to create ES256 token: %v", err)
		}

		_, err = VerifySupabaseJWT(tokenStr)
		if err == nil {
			t.Fatal("expected error for unknown kid, got nil")
		}
	})
}
