package classifier

import (
	"fmt"
	"net/mail"
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
	Reason     string
}

type Classifier struct{}

func New() *Classifier {
	return &Classifier{}
}

func (c *Classifier) Classify(message reader.Message, userRules []rules.Rule) Result {
	candidates := make([]candidateRule, 0, len(userRules))
	for _, rule := range userRules {
		if !rule.Enabled {
			continue
		}
		scope := normalizeScope(rule.Scope)
		if scope == "" {
			continue
		}
		candidates = append(candidates, candidateRule{Rule: rule, Scope: scope})
	}

	parts := extractMessageParts(message)
	var bestUserSpecific *matchResult
	var bestGlobal *matchResult
	for _, candidate := range candidates {
		match, ok := matchCandidate(parts, candidate)
		if !ok {
			continue
		}
		switch candidate.Scope {
		case rules.ScopeUserSpecific:
			if bestUserSpecific == nil || match.Score > bestUserSpecific.Score {
				copied := match
				bestUserSpecific = &copied
			}
		case rules.ScopeGlobal:
			if bestGlobal == nil || match.Score > bestGlobal.Score {
				copied := match
				bestGlobal = &copied
			}
		}
	}

	best := bestUserSpecific
	if best == nil {
		best = bestGlobal
	}
	if best == nil {
		return Result{
			Label:      LabelUnknown,
			Confidence: 0.5,
			Reason:     "no_matching_rule",
		}
	}

	return Result{
		Label:      best.TargetLabel,
		Confidence: confidenceFromScore(best.Score),
		Reason:     best.Reason,
	}
}

type candidateRule struct {
	Rule  rules.Rule
	Scope string
}

type messageParts struct {
	SenderEmail  string
	SenderDomain string
	Subject      string
	Body         string
	Any          string
}

type matchResult struct {
	TargetLabel string
	Score       int
	Reason      string
}

func extractMessageParts(message reader.Message) messageParts {
	email := extractSenderEmail(message.From)
	subject := normalize(message.Subject)
	body := normalize(message.BodySnippet)
	any := normalize(strings.Join([]string{message.From, message.Subject, message.BodySnippet}, " "))

	return messageParts{
		SenderEmail:  email,
		SenderDomain: extractSenderDomain(email),
		Subject:      subject,
		Body:         body,
		Any:          any,
	}
}

func matchCandidate(parts messageParts, candidate candidateRule) (matchResult, bool) {
	rule := candidate.Rule
	operator := normalize(rule.Operator)
	if operator == "" {
		operator = rules.OperatorContains
	}

	if rule.Priority < 0 {
		return matchResult{}, false
	}
	if strings.TrimSpace(rule.RuleValue) == "" || strings.TrimSpace(rule.TargetLabel) == "" {
		return matchResult{}, false
	}
	if operator != rules.OperatorContains && operator != rules.OperatorEquals {
		return matchResult{}, false
	}

	fieldValue := ""
	switch rule.RuleType {
	case rules.RuleTypeSenderEmail:
		fieldValue = parts.SenderEmail
	case rules.RuleTypeSenderDomain:
		fieldValue = parts.SenderDomain
	case rules.RuleTypeSubject:
		fieldValue = parts.Subject
	case rules.RuleTypeBody:
		fieldValue = parts.Body
	case rules.RuleTypeAny:
		fieldValue = parts.Any
	default:
		return matchResult{}, false
	}

	needle := normalize(rule.RuleValue)
	if needle == "" {
		return matchResult{}, false
	}

	var matched bool
	switch operator {
	case rules.OperatorEquals:
		matched = fieldValue == needle
	case rules.OperatorContains:
		matched = strings.Contains(fieldValue, needle)
	}
	if !matched {
		return matchResult{}, false
	}

	score := scoreRule(rule.Priority, specificityBonus(rule.RuleType, operator))
	return matchResult{
		TargetLabel: rule.TargetLabel,
		Score:       score,
		Reason: fmt.Sprintf(
			"scope=%s rule_type=%s operator=%s value=%s score=%d",
			candidate.Scope,
			rule.RuleType,
			operator,
			rule.RuleValue,
			score,
		),
	}, true
}

func scoreRule(priority int, specificity int) int {
	return priority*1000 + specificity
}

func specificityBonus(ruleType, operator string) int {
	operatorBonus := 0
	if operator == rules.OperatorEquals {
		operatorBonus = 20
	}

	typeBonus := 0
	switch ruleType {
	case rules.RuleTypeSenderEmail:
		typeBonus = 50
	case rules.RuleTypeSenderDomain:
		typeBonus = 40
	case rules.RuleTypeSubject:
		typeBonus = 30
	case rules.RuleTypeBody:
		typeBonus = 20
	case rules.RuleTypeAny:
		typeBonus = 10
	}

	return operatorBonus + typeBonus
}

func confidenceFromScore(score int) float64 {
	switch {
	case score >= 200000:
		return 0.98
	case score >= 120000:
		return 0.94
	default:
		return 0.9
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeScope(scope string) string {
	switch normalize(scope) {
	case "", rules.ScopeUserSpecific:
		return rules.ScopeUserSpecific
	case rules.ScopeGlobal:
		return rules.ScopeGlobal
	default:
		return ""
	}
}

func extractSenderEmail(from string) string {
	from = strings.TrimSpace(from)
	if from == "" {
		return ""
	}
	if parsed, err := mail.ParseAddress(from); err == nil {
		return normalize(parsed.Address)
	}
	return normalize(from)
}

func extractSenderDomain(senderEmail string) string {
	parts := strings.Split(senderEmail, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
