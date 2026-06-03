package services

import (
	"context"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

var resendClient *resend.Client

// InitEmail creates the Resend API client when RESEND_API_KEY is set.
func InitEmail() {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return
	}
	resendClient = resend.NewClient(apiKey)
}

// ResumeEmailHTML returns a simple branded HTML body for CV delivery emails.
func ResumeEmailHTML() string {
	return `<!DOCTYPE html>
<html>
<body style="font-family: system-ui, sans-serif; background:#131314; color:#e4e4e7; padding:32px;">
  <div style="max-width:480px; margin:0 auto;">
    <h1 style="color:#58c7ff; font-size:20px;">Your CV from Sethum</h1>
    <p style="line-height:1.6;">Thanks for your interest. My resume is attached to this email as a PDF.</p>
    <p style="line-height:1.6; color:#a1a1aa;">— Sethum Methsanda</p>
  </div>
</body>
</html>`
}

// SendEmail sends an HTML email with an optional PDF attachment via Resend.
func SendEmail(ctx context.Context, to, subject, bodyHTML string, attachment []byte, attachmentName string) error {
	if resendClient == nil {
		return fmt.Errorf("resend client not initialized")
	}

	from := os.Getenv("EMAIL_FROM")
	if from == "" {
		return fmt.Errorf("EMAIL_FROM is not set")
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    bodyHTML,
	}

	if len(attachment) > 0 && attachmentName != "" {
		params.Attachments = []*resend.Attachment{
			{
				Content:     attachment,
				Filename:    attachmentName,
				ContentType: "application/pdf",
			},
		}
	}

	_, err := resendClient.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}
	return nil
}
