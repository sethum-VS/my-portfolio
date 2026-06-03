package services

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func recaptchaMinScore() float32 {
	minScore := float32(0.5)
	if v := os.Getenv("RECAPTCHA_MIN_SCORE"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 32); err == nil {
			minScore = float32(parsed)
		}
	}
	return minScore
}

func requestClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if i := strings.IndexByte(forwarded, ','); i >= 0 {
			return strings.TrimSpace(forwarded[:i])
		}
		return strings.TrimSpace(forwarded)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// recaptchaScoreAcceptable decides whether a valid token passes score policy.
// Checkbox keys can return very low scores (or LOW_CONFIDENCE) even after the user
// passes the challenge; in those cases the token validity + PASSED challenge is authoritative.
func recaptchaScoreAcceptable(resp *recaptchaenterprisepb.Assessment, minScore float32) bool {
	if resp.RiskAnalysis == nil {
		return true
	}

	if resp.RiskAnalysis.Challenge == recaptchaenterprisepb.RiskAnalysis_PASSED {
		return true
	}

	for _, reason := range resp.RiskAnalysis.Reasons {
		if reason == recaptchaenterprisepb.RiskAnalysis_LOW_CONFIDENCE_SCORE {
			return true
		}
	}

	return resp.RiskAnalysis.Score >= minScore
}

// VerifyRecaptcha validates a reCAPTCHA Enterprise token and returns the risk score.
func VerifyRecaptcha(ctx context.Context, token string, r *http.Request) (float32, error) {
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

	event := &recaptchaenterprisepb.Event{
		Token:          token,
		SiteKey:        siteKey,
		ExpectedAction: recaptchaResumeAction,
	}
	if r != nil {
		event.UserAgent = r.UserAgent()
		event.UserIpAddress = requestClientIP(r)
	}

	req := &recaptchaenterprisepb.CreateAssessmentRequest{
		Parent: fmt.Sprintf("projects/%s", projectID),
		Assessment: &recaptchaenterprisepb.Assessment{
			Event: event,
		},
	}

	resp, err := client.CreateAssessment(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("create assessment: %w", err)
	}

	if resp.TokenProperties == nil || !resp.TokenProperties.Valid {
		reason := "unknown"
		if resp.TokenProperties != nil && resp.TokenProperties.InvalidReason != recaptchaenterprisepb.TokenProperties_INVALID_REASON_UNSPECIFIED {
			reason = resp.TokenProperties.InvalidReason.String()
		}
		return 0, fmt.Errorf("invalid recaptcha token (%s)", reason)
	}
	if action := resp.TokenProperties.Action; action != "" && action != recaptchaResumeAction {
		return 0, fmt.Errorf("invalid recaptcha action: got %q want %q", action, recaptchaResumeAction)
	}

	var score float32
	if resp.RiskAnalysis != nil {
		score = resp.RiskAnalysis.Score
	}

	minScore := recaptchaMinScore()
	if !recaptchaScoreAcceptable(resp, minScore) {
		return score, fmt.Errorf("recaptcha score %.2f below threshold %.2f", score, minScore)
	}

	return score, nil
}
