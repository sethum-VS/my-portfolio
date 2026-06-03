package models

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/iterator"
)

const (
	resumeConfigCollection = "resume_config"
	resumeConfigDocID      = "default"
	waitlistCollection     = "resume_waitlist"
)

// ResumeConfig holds global resume availability and storage location.
type ResumeConfig struct {
	IsComingSoon  bool   `firestore:"is_coming_soon"`
	PDFStorageURI string `firestore:"pdf_storage_uri"`
}

// GetResumeConfig returns the resume config document, or zero values if missing.
func GetResumeConfig() ResumeConfig {
	ctx := context.Background()
	if DB == nil {
		log.Println("Firestore client not initialized")
		return ResumeConfig{}
	}

	doc, err := DB.Collection(resumeConfigCollection).Doc(resumeConfigDocID).Get(ctx)
	if err != nil {
		return ResumeConfig{IsComingSoon: true}
	}

	var cfg ResumeConfig
	if err := doc.DataTo(&cfg); err != nil {
		log.Printf("Failed to map resume config: %v", err)
		return ResumeConfig{IsComingSoon: true}
	}
	return cfg
}

// SaveResumeConfig writes the resume config document.
func SaveResumeConfig(cfg ResumeConfig) error {
	ctx := context.Background()
	if DB == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	_, err := DB.Collection(resumeConfigCollection).Doc(resumeConfigDocID).Set(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to save resume config: %w", err)
	}
	return nil
}

// WaitlistCount returns how many emails are on the resume waitlist.
func WaitlistCount() int {
	return len(ListWaitlistEmails())
}

// ListWaitlistEmails returns all emails on the resume waitlist.
func ListWaitlistEmails() []string {
	ctx := context.Background()
	if DB == nil {
		log.Println("Firestore client not initialized")
		return nil
	}

	iter := DB.Collection(waitlistCollection).Documents(ctx)
	var emails []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Failed to iterate waitlist: %v", err)
			return emails
		}
		if email, ok := doc.Data()["email"].(string); ok && email != "" {
			emails = append(emails, email)
		}
	}
	return emails
}

// AddToWaitlist stores an email if not already present.
func AddToWaitlist(email string) error {
	ctx := context.Background()
	if DB == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return fmt.Errorf("email is required")
	}

	q := DB.Collection(waitlistCollection).Where("email", "==", email).Limit(1)
	docs, err := q.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to query waitlist: %w", err)
	}
	if len(docs) > 0 {
		return nil
	}

	_, err = DB.Collection(waitlistCollection).NewDoc().Set(ctx, map[string]interface{}{
		"email":      email,
		"created_at": time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to add to waitlist: %w", err)
	}
	return nil
}

// ClearWaitlist deletes all documents in the waitlist collection.
func ClearWaitlist() error {
	ctx := context.Background()
	if DB == nil {
		return fmt.Errorf("firestore client not initialized")
	}

	iter := DB.Collection(waitlistCollection).Documents(ctx)
	batch := DB.Batch()
	count := 0

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate waitlist: %w", err)
		}
		batch.Delete(doc.Ref)
		count++
		if count >= 400 {
			if _, err := batch.Commit(ctx); err != nil {
				return fmt.Errorf("failed to commit waitlist batch: %w", err)
			}
			batch = DB.Batch()
			count = 0
		}
	}

	if count > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit waitlist batch: %w", err)
		}
	}
	return nil
}
