package models

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ResumeConfig holds global resume availability and storage location.
type ResumeConfig struct {
	IsComingSoon  bool   `json:"is_coming_soon"`
	PDFStorageURI string `json:"pdf_storage_uri"`
}

// GetResumeConfig returns the resume config from PostgreSQL.
func GetResumeConfig(ctx context.Context) ResumeConfig {
	if err := checkDB(); err != nil {
		return ResumeConfig{IsComingSoon: true}
	}

	var cfg ResumeConfig
	query := `SELECT is_coming_soon, pdf_storage_uri FROM resume_config WHERE id = 'default'`
	err := DB.QueryRow(ctx, query).Scan(&cfg.IsComingSoon, &cfg.PDFStorageURI)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Printf("Failed to get resume config: %v", err)
		}
		return ResumeConfig{IsComingSoon: true}
	}
	return cfg
}

// SaveResumeConfig updates the resume config in PostgreSQL.
func SaveResumeConfig(ctx context.Context, cfg ResumeConfig) error {
	if err := checkDB(); err != nil {
		return err
	}

	query := `
		INSERT INTO resume_config (id, is_coming_soon, pdf_storage_uri)
		VALUES ('default', $1, $2)
		ON CONFLICT (id) DO UPDATE SET
			is_coming_soon = EXCLUDED.is_coming_soon,
			pdf_storage_uri = EXCLUDED.pdf_storage_uri`

	_, err := DB.Exec(ctx, query, cfg.IsComingSoon, cfg.PDFStorageURI)
	if err != nil {
		return fmt.Errorf("failed to save resume config: %w", err)
	}
	return nil
}

// WaitlistCount returns how many emails are on the resume waitlist.
func WaitlistCount(ctx context.Context) int {
	if err := checkDB(); err != nil {
		return 0
	}

	var count int
	err := DB.QueryRow(ctx, `SELECT COUNT(*) FROM resume_waitlist`).Scan(&count)
	if err != nil {
		log.Printf("Failed to count waitlist: %v", err)
		return 0
	}
	return count
}

// ListWaitlistEmails returns all emails on the resume waitlist.
func ListWaitlistEmails(ctx context.Context) []string {
	var emails []string
	if err := checkDB(); err != nil {
		return emails
	}

	rows, err := DB.Query(ctx, `SELECT email FROM resume_waitlist ORDER BY created_at ASC`)
	if err != nil {
		log.Printf("Failed to list waitlist: %v", err)
		return emails
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			log.Printf("Failed to scan email: %v", err)
			continue
		}
		if email != "" {
			emails = append(emails, email)
		}
	}
	return emails
}

// AddToWaitlist stores an email if not already present.
func AddToWaitlist(ctx context.Context, email string) error {
	if err := checkDB(); err != nil {
		return err
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("email is required")
	}

	query := `INSERT INTO resume_waitlist (email, created_at) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`
	_, err := DB.Exec(ctx, query, email, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to add to waitlist: %w", err)
	}
	return nil
}

// ClearWaitlist deletes all records in the resume_waitlist table.
func ClearWaitlist(ctx context.Context) error {
	if err := checkDB(); err != nil {
		return err
	}

	_, err := DB.Exec(ctx, `DELETE FROM resume_waitlist`)
	if err != nil {
		return fmt.Errorf("failed to clear waitlist: %w", err)
	}
	return nil
}
