package storage

import (
	"context"
	"database/sql"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

func (p *Postgres) ListEnabledUserRules(ctx context.Context, userID string) ([]models.UserRule, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, user_id, rule_type, operator, rule_value, target_label, enabled, priority
		 FROM user_rules
		 WHERE (user_id = $1 OR user_id IS NULL) AND enabled = TRUE
		 ORDER BY user_id IS NULL, priority DESC, id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.UserRule, 0)
	for rows.Next() {
		var rule models.UserRule
		var ruleUserID sql.NullString
		if err := rows.Scan(
			&rule.ID,
			&ruleUserID,
			&rule.RuleType,
			&rule.Operator,
			&rule.RuleValue,
			&rule.TargetLabel,
			&rule.Enabled,
			&rule.Priority,
		); err != nil {
			return nil, err
		}
		if ruleUserID.Valid {
			rule.UserID = &ruleUserID.String
		}
		result = append(result, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
