package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	DBPool         *pgxpool.Pool
	jwtSecret      []byte
	jwtSecretOnce  sync.Once
)

func InitSupabase(ctx context.Context) error {
	connStr := os.Getenv("SUPABASE_DB_URL")
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

func VerifySupabaseJWT(tokenStr string) (*jwt.Token, error) {
	jwtSecretOnce.Do(func() {
		jwtSecret = []byte(os.Getenv("SUPABASE_JWT_SECRET"))
	})

	if len(jwtSecret) == 0 {
		return nil, fmt.Errorf("SUPABASE_JWT_SECRET environment variable is not set")
	}

	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})
}
