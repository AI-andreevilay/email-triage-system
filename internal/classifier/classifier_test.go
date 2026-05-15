package classifier

import (
	"strings"
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
			if result.Reason == "" {
				t.Fatal("expected non-empty reason")
			}
		})
	}
}

func TestClassifierUserRulePreferredWhenPriorityEqual(t *testing.T) {
	c := New()

	message := reader.Message{
		From:        "newsletter@store.com",
		Subject:     "Weekend sale 50% off",
		BodySnippet: "Promo code inside. Limited offer.",
	}

	userRules := []rules.Rule{
		{
			RuleType:    rules.RuleTypeSenderEmail,
			Operator:    rules.OperatorEquals,
			RuleValue:   "newsletter@store.com",
			TargetLabel: LabelTransactions,
			Enabled:     true,
			Priority:    120,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelTransactions {
		t.Fatalf("label = %s, want %s", result.Label, LabelTransactions)
	}
	if !strings.Contains(result.Reason, "source=user") {
		t.Fatalf("reason = %s, expected user source", result.Reason)
	}
}

func TestClassifierHigherPriorityWinsRegardlessOfSource(t *testing.T) {
	c := New()

	message := reader.Message{
		From:        "alerts@bank.com",
		Subject:     "card transaction approved",
		BodySnippet: "Paid successfully.",
	}

	userRules := []rules.Rule{
		{
			RuleType:    rules.RuleTypeSenderEmail,
			Operator:    rules.OperatorEquals,
			RuleValue:   "alerts@bank.com",
			TargetLabel: LabelSocial,
			Enabled:     true,
			Priority:    50,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelTransactions {
		t.Fatalf("label = %s, want %s", result.Label, LabelTransactions)
	}
}

func TestClassifierSpecificityWhenPriorityAndSourceEqual(t *testing.T) {
	c := New()

	message := reader.Message{
		From:        "no-reply@accounts.google.com",
		Subject:     "Security update",
		BodySnippet: "Please review new sign-in",
	}

	userRules := []rules.Rule{
		{
			RuleType:    rules.RuleTypeSenderDomain,
			Operator:    rules.OperatorContains,
			RuleValue:   "google.com",
			TargetLabel: LabelPromo,
			Enabled:     true,
			Priority:    100,
		},
		{
			RuleType:    rules.RuleTypeSenderEmail,
			Operator:    rules.OperatorEquals,
			RuleValue:   "no-reply@accounts.google.com",
			TargetLabel: LabelSecurity,
			Enabled:     true,
			Priority:    100,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelSecurity {
		t.Fatalf("label = %s, want %s", result.Label, LabelSecurity)
	}
}

func TestClassifierRealWorldJobMails(t *testing.T) {
	c := New()

	tests := []struct {
		name      string
		message   reader.Message
		wantLabel string
	}{
		{
			name: "application received",
			message: reader.Message{
				From:        "no-reply@chronosphere.io",
				Subject:     "Thanks for applying to Chronosphere",
				BodySnippet: "Your application has been received and we will review it right away.",
			},
			wantLabel: LabelJob,
		},
		{
			name: "application for backend engineer",
			message: reader.Message{
				From:        "careers@example.com",
				Subject:     "Application for Backend Engineer (Go-lang)",
				BodySnippet: "We have received your application and will contact you with status.",
			},
			wantLabel: LabelJob,
		},
		{
			name: "subscription upsell",
			message: reader.Message{
				From:        "no-reply@freelancer.com",
				Subject:     "Your Freelancer Plus subscription will end soon",
				BodySnippet: "Take action now to keep your tools and upgrade your plan.",
			},
			wantLabel: LabelPromo,
		},
		{
			name: "hiring test invitation",
			message: reader.Message{
				From:        "recruiting@testgorilla.com",
				Subject:     "Your application is live",
				BodySnippet: "Complete your assigned tests so we can match you with your ideal role.",
			},
			wantLabel: LabelJob,
		},
		{
			name: "rejection email from hiring team",
			message: reader.Message{
				From:        "hiring@glacis.com",
				Subject:     "Update on your application",
				BodySnippet: "Glacis Hiring Team reviewed your application and will not move forward.",
			},
			wantLabel: LabelJob,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Classify(tt.message, nil)
			if result.Label != tt.wantLabel {
				t.Fatalf("label = %s, want %s, reason=%s", result.Label, tt.wantLabel, result.Reason)
			}
		})
	}
}
