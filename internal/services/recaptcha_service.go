package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"

	recaptchaenterprise "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchaenterprisepb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

const recaptchaResumeAction = "resume_request"

var (
	recaptchaClient     *recaptchaenterprise.Client
	recaptchaClientOnce sync.Once
	recaptchaClientErr  error
)

func getRecaptchaClient(ctx context.Context) (*recaptchaenterprise.Client, error) {
	recaptchaClientOnce.Do(func() {
		recaptchaClient, recaptchaClientErr = recaptchaenterprise.NewClient(ctx)
	})
	return recaptchaClient, recaptchaClientErr
}

// VerifyRecaptcha validates a reCAPTCHA Enterprise token and returns the risk score.
func VerifyRecaptcha(ctx context.Context, token string) (float32, error) {
	if token == "" {
		return 0, fmt.Errorf("missing recaptcha token")
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	siteKey := os.Getenv("RECAPTCHA_SITE_KEY")
	if projectID == "" || siteKey == "" {
		return 0, fmt.Errorf("recaptcha is not configured")
	}

	client, err := getRecaptchaClient(ctx)
	if err != nil {
		return 0, fmt.Errorf("recaptcha client: %w", err)
	}

	req := &recaptchaenterprisepb.CreateAssessmentRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Assessment: &recaptchaenterprisepb.Assessment{
			Event: &recaptchaenterprisepb.Event{
				Token:   token,
				SiteKey: siteKey,
			},
		},
	}

	resp, err := client.CreateAssessment(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("create assessment: %w", err)
	}

	if resp.TokenProperties == nil || !resp.TokenProperties.Valid {
		return 0, fmt.Errorf("invalid recaptcha token")
	}
	if action := resp.TokenProperties.Action; action != "" && action != recaptchaResumeAction {
		return 0, fmt.Errorf("invalid recaptcha action")
	}

	var score float32
	if resp.RiskAnalysis != nil {
		score = resp.RiskAnalysis.Score
	}

	minScore := float32(0.5)
	if v := os.Getenv("RECAPTCHA_MIN_SCORE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 32); err == nil {
			minScore = float32(parsed)
		}
	}

	if score < minScore {
		return score, fmt.Errorf("recaptcha score below threshold")
	}

	return score, nil
}
