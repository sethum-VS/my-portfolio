package services

import (
	"github.com/sethum-VS/my-portfolio/internal/config"

	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JWK struct {
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

var (
	DBPool        *pgxpool.Pool
	jwtSecret     []byte
	jwtSecretOnce sync.Once

	jwkCache       map[string]*ecdsa.PublicKey
	jwkCacheMu     sync.RWMutex
	jwkCacheExpiry time.Time
)

func InitSupabase(ctx context.Context) error {
	connStr := config.AppConfig.SupabaseDBURL
	if connStr == "" {
		return fmt.Errorf("SUPABASE_DB_URL is not set")
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure pool settings
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DBPool = pool
	log.Println("✓ Supabase PostgreSQL connection pool initialized successfully")
	return nil
}

func getJWKPublicKey(kid string) (*ecdsa.PublicKey, error) {
	jwkCacheMu.RLock()
	pubKey, exists := jwkCache[kid]
	expired := time.Now().After(jwkCacheExpiry)
	jwkCacheMu.RUnlock()

	if exists && !expired {
		return pubKey, nil
	}

	// Fetch JWKS
	supabaseURL := config.AppConfig.SupabaseURL
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL environment variable is not set")
	}

	url := fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL)
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JWKS request returned status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	jwkCacheMu.Lock()
	defer jwkCacheMu.Unlock()

	// Clear and rebuild cache
	jwkCache = make(map[string]*ecdsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Alg != "ES256" || key.Kty != "EC" {
			continue
		}
		xBytes, err := base64.RawURLEncoding.DecodeString(key.X)
		if err != nil {
			continue
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(key.Y)
		if err != nil {
			continue
		}
		ecdsaKey := &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}
		jwkCache[key.Kid] = ecdsaKey
	}

	jwkCacheExpiry = time.Now().Add(24 * time.Hour)

	pubKey, exists = jwkCache[kid]
	if !exists {
		return nil, fmt.Errorf("JWK with kid %s not found in JWKS", kid)
	}

	return pubKey, nil
}

func VerifySupabaseJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		alg, ok := token.Header["alg"].(string)
		if !ok {
			return nil, fmt.Errorf("missing alg header")
		}

		switch alg {
		case "HS256":
			jwtSecretOnce.Do(func() {
				jwtSecret = []byte(config.AppConfig.SupabaseJWTSecret)
			})
			if len(jwtSecret) == 0 {
				return nil, fmt.Errorf("SUPABASE_JWT_SECRET environment variable is not set")
			}
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method for HS256: %v", token.Header["alg"])
			}
			return jwtSecret, nil

		case "ES256":
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, fmt.Errorf("missing kid header for ES256")
			}
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("unexpected signing method for ES256: %v", token.Header["alg"])
			}
			return getJWKPublicKey(kid)

		default:
			return nil, fmt.Errorf("unsupported signing algorithm: %s", alg)
		}
	})
}
