package services

import (
	"testing"

	recaptchaenterprisepb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

func TestRecaptchaScoreAcceptable(t *testing.T) {
	t.Parallel()

	minScore := float32(0.5)

	t.Run("nil risk analysis", func(t *testing.T) {
		if !recaptchaScoreAcceptable(&recaptchaenterprisepb.Assessment{}, minScore) {
			t.Fatal("expected acceptable when risk analysis is nil")
		}
	})

	t.Run("checkbox passed", func(t *testing.T) {
		resp := &recaptchaenterprisepb.Assessment{
			RiskAnalysis: &recaptchaenterprisepb.RiskAnalysis{
				Score:     0.0,
				Challenge: recaptchaenterprisepb.RiskAnalysis_PASSED,
			},
		}
		if !recaptchaScoreAcceptable(resp, minScore) {
			t.Fatal("expected acceptable when challenge passed")
		}
	})

	t.Run("low confidence", func(t *testing.T) {
		resp := &recaptchaenterprisepb.Assessment{
			RiskAnalysis: &recaptchaenterprisepb.RiskAnalysis{
				Score:   0.0,
				Reasons: []recaptchaenterprisepb.RiskAnalysis_ClassificationReason{recaptchaenterprisepb.RiskAnalysis_LOW_CONFIDENCE_SCORE},
			},
		}
		if !recaptchaScoreAcceptable(resp, minScore) {
			t.Fatal("expected acceptable for low confidence score")
		}
	})

	t.Run("score below threshold", func(t *testing.T) {
		resp := &recaptchaenterprisepb.Assessment{
			RiskAnalysis: &recaptchaenterprisepb.RiskAnalysis{Score: 0.2},
		}
		if recaptchaScoreAcceptable(resp, minScore) {
			t.Fatal("expected rejection for low score")
		}
	})

	t.Run("score at threshold", func(t *testing.T) {
		resp := &recaptchaenterprisepb.Assessment{
			RiskAnalysis: &recaptchaenterprisepb.RiskAnalysis{Score: 0.5},
		}
		if !recaptchaScoreAcceptable(resp, minScore) {
			t.Fatal("expected acceptable at threshold")
		}
	})
}
