package rules

const (
	RuleTypeSenderEmail  = "sender_email"
	RuleTypeSenderDomain = "sender_domain"
	RuleTypeSubject      = "subject"
	RuleTypeBody         = "body"
	RuleTypeAny          = "any"
)

const (
	OperatorEquals   = "equals"
	OperatorContains = "contains"
)

type Rule struct {
	RuleType    string
	Operator    string
	RuleValue   string
	TargetLabel string
	Enabled     bool
	Priority    int
}
