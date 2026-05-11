package classifier

import (
	"slices"
	"strings"

	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/rules"
)

const (
	LabelJob          = "Job"
	LabelTransactions = "Transactions"
	LabelSecurity     = "Security"
	LabelPromo        = "Promo"
	LabelSocial       = "Social"
	LabelUnknown      = "Unknown"
)

type Result struct {
	Label      string
	Confidence float64
}

type Classifier struct{}

func New() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Classify(message reader.Message, userRules []rules.Rule) Result {
	if result, ok := matchUserRules(message, userRules); ok {
		return result
	}

	text := normalize(strings.Join([]string{message.From, message.Subject, message.BodySnippet}, " "))
	switch {
	case containsAny(text, securityKeywords):
		return Result{Label: LabelSecurity, Confidence: 0.94}
	case containsAny(text, transactionKeywords):
		return Result{Label: LabelTransactions, Confidence: 0.92}
	case containsAny(text, jobKeywords):
		return Result{Label: LabelJob, Confidence: 0.9}
	case containsAny(text, promoKeywords):
		return Result{Label: LabelPromo, Confidence: 0.88}
	case containsAny(text, socialKeywords):
		return Result{Label: LabelSocial, Confidence: 0.86}
	default:
		return Result{Label: LabelUnknown, Confidence: 0.5}
	}
}

func matchUserRules(message reader.Message, userRules []rules.Rule) (Result, bool) {
	activeRules := make([]rules.Rule, 0, len(userRules))
	for _, rule := range userRules {
		if !rule.Enabled {
			continue
		}
		if strings.TrimSpace(rule.RuleValue) == "" || strings.TrimSpace(rule.TargetLabel) == "" {
			continue
		}
		activeRules = append(activeRules, rule)
	}

	slices.SortFunc(activeRules, func(a, b rules.Rule) int {
		if a.Priority < b.Priority {
			return -1
		}
		if a.Priority > b.Priority {
			return 1
		}
		return 0
	})

	from := normalize(message.From)
	subject := normalize(message.Subject)
	body := normalize(message.BodySnippet)
	anyText := normalize(strings.Join([]string{message.From, message.Subject, message.BodySnippet}, " "))

	for _, rule := range activeRules {
		value := normalize(rule.RuleValue)
		if value == "" {
			continue
		}
		matched := false
		switch rule.RuleType {
		case rules.RuleTypeFromContains:
			matched = strings.Contains(from, value)
		case rules.RuleTypeSubjectContains:
			matched = strings.Contains(subject, value)
		case rules.RuleTypeBodyContains:
			matched = strings.Contains(body, value)
		case rules.RuleTypeAnyContains:
			matched = strings.Contains(anyText, value)
		default:
			continue
		}
		if matched {
			return Result{
				Label:      rule.TargetLabel,
				Confidence: 0.99,
			}, true
		}
	}

	return Result{}, false
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

var securityKeywords = []string{
	"security",
	"sign-in",
	"signin",
	"password",
	"otp",
	"verification code",
	"2fa",
	"suspicious login",
}

var transactionKeywords = []string{
	"transaction",
	"payment",
	"invoice",
	"receipt",
	"charged",
	"purchase",
	"order",
	"statement",
	"card",
	"qr pay",
}

var jobKeywords = []string{
	"job",
	"interview",
	"recruiter",
	"opportunity",
	"vacancy",
	"resume",
}

var promoKeywords = []string{
	"sale",
	"discount",
	"promo",
	"coupon",
	"offer",
	"deal",
	"% off",
}

var socialKeywords = []string{
	"notification",
	"friend",
	"comment",
	"like",
	"mention",
	"invitation",
}
