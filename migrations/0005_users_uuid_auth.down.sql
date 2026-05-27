DROP INDEX IF EXISTS email_messages_user_id_idx;
DROP INDEX IF EXISTS scan_runs_user_id_idx;
DROP INDEX IF EXISTS user_rules_unique_rule;

ALTER TABLE user_rules
    ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;

ALTER TABLE email_messages
    ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;

ALTER TABLE scan_runs
    ALTER COLUMN user_id TYPE TEXT USING user_id::TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS user_rules_unique_rule
    ON user_rules (user_id, rule_type, operator, rule_value, target_label)
    NULLS NOT DISTINCT;

DROP TABLE IF EXISTS users;
