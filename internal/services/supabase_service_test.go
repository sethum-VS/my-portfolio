package services

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func TestVerifySupabaseJWT(t *testing.T) {
	// Initialize the secret before running VerifySupabaseJWT
	secret := "test-jwt-secret-key-with-sufficient-length"
	os.Setenv("SUPABASE_JWT_SECRET", secret)
	defer os.Unsetenv("SUPABASE_JWT_SECRET")

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
