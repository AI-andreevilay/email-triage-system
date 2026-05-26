package models

type UserRule struct {
	ID          int64
	UserID      *string
	RuleType    string
	Operator    string
	RuleValue   string
	TargetLabel string
	Enabled     bool
	Priority    int
}
