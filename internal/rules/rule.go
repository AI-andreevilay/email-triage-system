package rules

const (
	RuleTypeFromContains    = "from_contains"
	RuleTypeSubjectContains = "subject_contains"
	RuleTypeBodyContains    = "body_contains"
	RuleTypeAnyContains     = "any_contains"
)

type Rule struct {
	RuleType    string
	RuleValue   string
	TargetLabel string
	Enabled     bool
	Priority    int
}
