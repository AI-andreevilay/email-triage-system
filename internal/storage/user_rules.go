package storage

import (
	"context"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

func (p *Postgres) ListEnabledUserRules(ctx context.Context, userID string) ([]models.UserRule, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT id, user_id, rule_type, rule_value, target_label, enabled, priority
		 FROM user_rules
		 WHERE user_id = $1 AND enabled = TRUE
		 ORDER BY priority ASC, id ASC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]models.UserRule, 0)
	for rows.Next() {
		var rule models.UserRule
		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.RuleType,
			&rule.RuleValue,
			&rule.TargetLabel,
			&rule.Enabled,
			&rule.Priority,
		); err != nil {
			return nil, err
		}
		result = append(result, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
