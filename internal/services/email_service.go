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

// Resume email subject options (pick one via ResumeEmailSubject or change the default below):
//
//  1. "Sethum Methsanda · Software Developer — resume attached"
//  2. "Thanks for your interest — my CV is inside"
//  3. "Resume ready · Sethum Methsanda"
//  4. "From sethum.dev — your requested resume"
const resumeEmailSubject = "Sethum Methsanda · Software Developer — resume attached"

// ResumeEmailSubject returns the subject line for CV delivery emails.
func ResumeEmailSubject() string {
	return resumeEmailSubject
}

// ResumeEmailHTML returns a branded HTML body for CV delivery emails (inline styles for Gmail/Outlook).
func ResumeEmailHTML() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="IE=edge">
  <title>Resume from Sethum Methsanda</title>
</head>
<body style="margin:0; padding:0; width:100%; background-color:#131314; background-image:radial-gradient(ellipse 80% 50% at 50% -10%, #252528 0%, #131314 65%); font-family:-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-text-size-adjust:100%; -ms-text-size-adjust:100%;">
  <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="background-color:#131314; background-image:radial-gradient(ellipse 80% 50% at 50% -10%, #252528 0%, #131314 65%);">
    <tr>
      <td align="center" style="padding:40px 16px;">
        <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="max-width:600px; background-color:#1e1e20; border:1px solid rgba(255,255,255,0.1); border-radius:16px;">
          <tr>
            <td style="padding:40px 32px 32px 32px;">
              <p style="margin:0 0 8px 0; font-size:11px; font-weight:600; letter-spacing:0.12em; text-transform:uppercase; color:#58c7ff;">Portfolio</p>
              <h1 style="margin:0 0 24px 0; font-size:22px; font-weight:600; line-height:1.35; color:#ffffff;">Sethum Methsanda <span style="color:#52525b;">//</span> Software Engineer</h1>
              <p style="margin:0 0 20px 0; font-size:16px; line-height:1.65; color:#a1a1aa;">Thank you for taking the time to reach out and for your interest in my work. I genuinely appreciate you considering me — it means a lot.</p>
              <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="margin:0 0 28px 0;">
                <tr>
                  <td style="padding:16px 20px; background-color:rgba(255,255,255,0.04); border:1px solid rgba(255,255,255,0.08); border-left:3px solid #58c7ff; border-radius:12px;">
                    <p style="margin:0 0 4px 0; font-size:14px; font-weight:600; color:#ffffff;">Resume attached</p>
                    <p style="margin:0; font-size:14px; line-height:1.55; color:#a1a1aa;">My CV is included with this email as a PDF — ready for you to review at your convenience.</p>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 32px 0; font-size:15px; line-height:1.6; color:#a1a1aa;">If you have any questions or would like to connect further, feel free to reply to this email or find me through the links below.</p>
              <p style="margin:0 0 20px 0; font-size:15px; line-height:1.5; color:#ffffff;">Warm regards,<br><span style="color:#a1a1aa;">Sethum Methsanda</span></p>
            </td>
          </tr>
          <tr>
            <td style="padding:0 32px 32px 32px;">
              <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%" style="border-top:1px solid rgba(255,255,255,0.08);">
                <tr>
                  <td style="padding-top:24px;">
                    <p style="margin:0 0 12px 0; font-size:12px; letter-spacing:0.08em; text-transform:uppercase; color:#71717a;">Connect</p>
                    <p style="margin:0; font-size:14px; line-height:1.8;">
                      <a href="https://sethum.dev" style="color:#58c7ff; text-decoration:none;">sethum.dev</a>
                      <span style="color:#3f3f46;">&nbsp;&middot;&nbsp;</span>
                      <a href="https://github.com/sethum-VS/" style="color:#58c7ff; text-decoration:none;">GitHub</a>
                      <span style="color:#3f3f46;">&nbsp;&middot;&nbsp;</span>
                      <a href="https://www.linkedin.com/in/sethumm" style="color:#58c7ff; text-decoration:none;">LinkedIn</a>
                    </p>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
        </table>
        <p style="margin:24px 0 0 0; font-size:12px; color:#52525b; text-align:center;">You received this because you requested a resume from sethum.dev</p>
      </td>
    </tr>
  </table>
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
