package classifier

import (
	"testing"

	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/rules"
)

func TestClassifierDefaultRules(t *testing.T) {
	c := New()

	tests := []struct {
		name      string
		message   reader.Message
		wantLabel string
	}{
		{
			name: "job",
			message: reader.Message{
				From:        "jobs@company.com",
				Subject:     "Backend Go Engineer opportunity",
				BodySnippet: "We reviewed your profile and would like to schedule an interview.",
			},
			wantLabel: LabelJob,
		},
		{
			name: "transactions",
			message: reader.Message{
				From:        "alerts@bank.com",
				Subject:     "Your card transaction was approved",
				BodySnippet: "Transaction amount $42.15 at Coffee Shop.",
			},
			wantLabel: LabelTransactions,
		},
		{
			name: "security",
			message: reader.Message{
				From:        "security@google.com",
				Subject:     "New sign-in detected",
				BodySnippet: "We noticed a new sign-in to your account.",
			},
			wantLabel: LabelSecurity,
		},
		{
			name: "promo",
			message: reader.Message{
				From:        "newsletter@store.com",
				Subject:     "Weekend sale 50% off",
				BodySnippet: "Promo code inside. Limited offer.",
			},
			wantLabel: LabelPromo,
		},
		{
			name: "social",
			message: reader.Message{
				From:        "friend@social.app",
				Subject:     "You have new notifications",
				BodySnippet: "5 new likes and 2 comments.",
			},
			wantLabel: LabelSocial,
		},
		{
			name: "unknown",
			message: reader.Message{
				From:        "someone@example.com",
				Subject:     "Meeting follow-up",
				BodySnippet: "Let's sync tomorrow.",
			},
			wantLabel: LabelUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Classify(tt.message, nil)
			if result.Label != tt.wantLabel {
				t.Fatalf("label = %s, want %s", result.Label, tt.wantLabel)
			}
		})
	}
}

func TestClassifierUserRulesPriority(t *testing.T) {
	c := New()

	message := reader.Message{
		From:        "newsletter@store.com",
		Subject:     "Weekend sale 50% off",
		BodySnippet: "Promo code inside. Limited offer.",
	}

	userRules := []rules.Rule{
		{
			RuleType:    rules.RuleTypeAnyContains,
			RuleValue:   "sale",
			TargetLabel: LabelSocial,
			Enabled:     true,
			Priority:    200,
		},
		{
			RuleType:    rules.RuleTypeFromContains,
			RuleValue:   "newsletter@store.com",
			TargetLabel: LabelTransactions,
			Enabled:     true,
			Priority:    100,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelTransactions {
		t.Fatalf("label = %s, want %s", result.Label, LabelTransactions)
	}
	if result.Confidence != 0.99 {
		t.Fatalf("confidence = %v, want 0.99", result.Confidence)
	}
}
