package services

import (
	"context"
	"fmt"
	"log"

	"github.com/sethum-VS/my-portfolio/internal/models"
)

const resumeAttachmentName = "Seth_Ummethsanda_CV.pdf"

// BroadcastResumeToWaitlist emails the current PDF to all waitlisted addresses and clears the list.
func BroadcastResumeToWaitlist(ctx context.Context, pdfURI string) {
	if err := broadcastResumeToWaitlist(ctx, pdfURI); err != nil {
		log.Printf("waitlist broadcast error: %v", err)
	}
}

func broadcastResumeToWaitlist(ctx context.Context, pdfURI string) error {
	emails := models.ListWaitlistEmails()
	if len(emails) == 0 {
		return nil
	}

	pdfBytes, err := DownloadResumePDF(ctx, pdfURI)
	if err != nil {
		return err
	}

	subject := ResumeEmailSubject()
	body := ResumeEmailHTML()

	var failed int
	for _, email := range emails {
		if err := SendEmail(ctx, email, subject, body, pdfBytes, resumeAttachmentName); err != nil {
			log.Printf("waitlist email failed for %s: %v", email, err)
			failed++
		}
	}

	if failed > 0 {
		log.Printf("waitlist broadcast: %d of %d emails failed", failed, len(emails))
		return fmt.Errorf("waitlist broadcast: %d of %d emails failed", failed, len(emails))
	}

	if err := models.ClearWaitlist(); err != nil {
		return err
	}
	return nil
}
