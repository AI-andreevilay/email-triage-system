package classifier

import (
	"strings"
	"testing"

	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/rules"
)

func TestClassifierReturnsUnknownWithoutRules(t *testing.T) {
	c := New()

	result := c.Classify(reader.Message{
		From:        "security@google.com",
		Subject:     "New sign-in detected",
		BodySnippet: "We noticed a new sign-in to your account.",
	}, nil)
	if result.Label != LabelUnknown {
		t.Fatalf("label = %s, want %s", result.Label, LabelUnknown)
	}
	if result.Reason != "no_matching_rule" {
		t.Fatalf("reason = %s, want no_matching_rule", result.Reason)
	}
}

func TestClassifierGlobalRules(t *testing.T) {
	c := New()
	globalRules := seededGlobalRulesForTest()

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
			result := c.Classify(tt.message, globalRules)
			if result.Label != tt.wantLabel {
				t.Fatalf("label = %s, want %s", result.Label, tt.wantLabel)
			}
			if result.Reason == "" {
				t.Fatal("expected non-empty reason")
			}
		})
	}
}

func TestClassifierUserSpecificRulePreferredOverGlobalRule(t *testing.T) {
	c := New()

	message := reader.Message{
		From:        "newsletter@store.com",
		Subject:     "Weekend sale 50% off",
		BodySnippet: "Promo code inside. Limited offer.",
	}

	userRules := []rules.Rule{
		{
			RuleType:    rules.RuleTypeAny,
			Operator:    rules.OperatorContains,
			RuleValue:   "sale",
			TargetLabel: LabelPromo,
			Enabled:     true,
			Priority:    999,
			Scope:       rules.ScopeGlobal,
		},
		{
			RuleType:    rules.RuleTypeSenderEmail,
			Operator:    rules.OperatorEquals,
			RuleValue:   "newsletter@store.com",
			TargetLabel: LabelTransactions,
			Enabled:     true,
			Priority:    1,
			Scope:       rules.ScopeUserSpecific,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelTransactions {
		t.Fatalf("label = %s, want %s", result.Label, LabelTransactions)
	}
	if !strings.Contains(result.Reason, "scope=user_specific") {
		t.Fatalf("reason = %s, expected user-specific scope", result.Reason)
	}
}

func TestClassifierHigherPriorityWinsWithinSameScope(t *testing.T) {
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
			Scope:       rules.ScopeGlobal,
		},
		{
			RuleType:    rules.RuleTypeSenderDomain,
			Operator:    rules.OperatorContains,
			RuleValue:   "bank.com",
			TargetLabel: LabelTransactions,
			Enabled:     true,
			Priority:    180,
			Scope:       rules.ScopeGlobal,
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
			Scope:       rules.ScopeUserSpecific,
		},
		{
			RuleType:    rules.RuleTypeSenderEmail,
			Operator:    rules.OperatorEquals,
			RuleValue:   "no-reply@accounts.google.com",
			TargetLabel: LabelSecurity,
			Enabled:     true,
			Priority:    100,
			Scope:       rules.ScopeUserSpecific,
		},
	}

	result := c.Classify(message, userRules)
	if result.Label != LabelSecurity {
		t.Fatalf("label = %s, want %s", result.Label, LabelSecurity)
	}
}

func TestClassifierRealWorldJobMails(t *testing.T) {
	c := New()
	globalRules := seededGlobalRulesForTest()
	userRules := append([]rules.Rule{}, globalRules...)
	userRules = append(userRules, rules.Rule{
		RuleType:    rules.RuleTypeSenderEmail,
		Operator:    rules.OperatorEquals,
		RuleValue:   "contact.center@permatabank.co.id",
		TargetLabel: LabelTransactions,
		Enabled:     true,
		Priority:    260,
		Scope:       rules.ScopeUserSpecific,
	})

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
		{
			name: "permatabank transaction sender",
			message: reader.Message{
				From:        "contact.center@permatabank.co.id",
				Subject:     "Informasi Transaksi",
				BodySnippet: "Terima kasih telah menggunakan layanan kami.",
			},
			wantLabel: LabelTransactions,
		},
		{
			name: "steam guard authenticator security",
			message: reader.Message{
				From:        "noreply@steampowered.com",
				Subject:     "Steam Guard Mobile Authenticator",
				BodySnippet: "An SMS code has been sent to remove or replace the Steam Guard Mobile Authenticator on your account.",
			},
			wantLabel: LabelSecurity,
		},
		{
			name: "job rejection talent acquisition",
			message: reader.Message{
				From:        "talent@cozey.com",
				Subject:     "Update on your application",
				BodySnippet: "Thank you for taking the time to apply. We will not be proceeding at this time. The Talent Acquisition Team.",
			},
			wantLabel: LabelJob,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Classify(tt.message, userRules)
			if result.Label != tt.wantLabel {
				t.Fatalf("label = %s, want %s, reason=%s", result.Label, tt.wantLabel, result.Reason)
			}
		})
	}
}

func seededGlobalRulesForTest() []rules.Rule {
	return []rules.Rule{
		{RuleType: rules.RuleTypeSenderDomain, Operator: rules.OperatorContains, RuleValue: "google.com", TargetLabel: LabelSecurity, Enabled: true, Priority: 220, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeSubject, Operator: rules.OperatorContains, RuleValue: "sign-in", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "verification code", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "otp", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "2fa", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "authenticator", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "lock your account", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "если это не были вы", TargetLabel: LabelSecurity, Enabled: true, Priority: 210, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "вошла в систему", TargetLabel: LabelSecurity, Enabled: true, Priority: 205, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "подтвердить адрес эл. почты", TargetLabel: LabelSecurity, Enabled: true, Priority: 205, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeSenderDomain, Operator: rules.OperatorContains, RuleValue: "bank.com", TargetLabel: LabelTransactions, Enabled: true, Priority: 180, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "transaction", TargetLabel: LabelTransactions, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "payment", TargetLabel: LabelTransactions, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "invoice", TargetLabel: LabelTransactions, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "receipt", TargetLabel: LabelTransactions, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "card", TargetLabel: LabelTransactions, Enabled: true, Priority: 160, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "qr pay", TargetLabel: LabelTransactions, Enabled: true, Priority: 160, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "thank you for applying", TargetLabel: LabelJob, Enabled: true, Priority: 180, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "your application has been received", TargetLabel: LabelJob, Enabled: true, Priority: 180, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "we have received your application", TargetLabel: LabelJob, Enabled: true, Priority: 180, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "thank you for taking the time to apply", TargetLabel: LabelJob, Enabled: true, Priority: 180, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "talent acquisition team", TargetLabel: LabelJob, Enabled: true, Priority: 175, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "will not be proceeding at this time", TargetLabel: LabelJob, Enabled: true, Priority: 175, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "hiring team", TargetLabel: LabelJob, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "backend engineer", TargetLabel: LabelJob, Enabled: true, Priority: 170, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "applicant", TargetLabel: LabelJob, Enabled: true, Priority: 165, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "position", TargetLabel: LabelJob, Enabled: true, Priority: 155, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "role", TargetLabel: LabelJob, Enabled: true, Priority: 150, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "interview", TargetLabel: LabelJob, Enabled: true, Priority: 150, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "recruiter", TargetLabel: LabelJob, Enabled: true, Priority: 150, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "job", TargetLabel: LabelJob, Enabled: true, Priority: 140, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "opportunity", TargetLabel: LabelJob, Enabled: true, Priority: 140, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "vacancy", TargetLabel: LabelJob, Enabled: true, Priority: 140, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "resume", TargetLabel: LabelJob, Enabled: true, Priority: 130, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "subscription will end", TargetLabel: LabelPromo, Enabled: true, Priority: 145, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "renew", TargetLabel: LabelPromo, Enabled: true, Priority: 140, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "upgrade", TargetLabel: LabelPromo, Enabled: true, Priority: 140, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "keep your", TargetLabel: LabelPromo, Enabled: true, Priority: 130, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "sale", TargetLabel: LabelPromo, Enabled: true, Priority: 120, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "discount", TargetLabel: LabelPromo, Enabled: true, Priority: 120, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "promo", TargetLabel: LabelPromo, Enabled: true, Priority: 120, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "coupon", TargetLabel: LabelPromo, Enabled: true, Priority: 120, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "offer", TargetLabel: LabelPromo, Enabled: true, Priority: 110, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "deal", TargetLabel: LabelPromo, Enabled: true, Priority: 110, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "% off", TargetLabel: LabelPromo, Enabled: true, Priority: 110, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "notification", TargetLabel: LabelSocial, Enabled: true, Priority: 100, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "friend", TargetLabel: LabelSocial, Enabled: true, Priority: 100, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "comment", TargetLabel: LabelSocial, Enabled: true, Priority: 100, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "mention", TargetLabel: LabelSocial, Enabled: true, Priority: 100, Scope: rules.ScopeGlobal},
		{RuleType: rules.RuleTypeAny, Operator: rules.OperatorContains, RuleValue: "invitation", TargetLabel: LabelSocial, Enabled: true, Priority: 100, Scope: rules.ScopeGlobal},
	}
}
