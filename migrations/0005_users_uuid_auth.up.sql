CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT UNIQUE NOT NULL,
    telegram_username TEXT,
    display_name TEXT,
    role TEXT NOT NULL CHECK (role IN ('user', 'admin')),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE scan_runs
    ALTER COLUMN user_id TYPE UUID USING user_id::UUID;

ALTER TABLE email_messages
    ALTER COLUMN user_id TYPE UUID USING user_id::UUID;

DROP INDEX IF EXISTS user_rules_unique_rule;

ALTER TABLE user_rules
    ALTER COLUMN user_id TYPE UUID USING user_id::UUID;

CREATE UNIQUE INDEX IF NOT EXISTS user_rules_unique_rule
    ON user_rules (user_id, rule_type, operator, rule_value, target_label)
    NULLS NOT DISTINCT;

CREATE INDEX IF NOT EXISTS scan_runs_user_id_idx
    ON scan_runs (user_id);

CREATE INDEX IF NOT EXISTS email_messages_user_id_idx
    ON email_messages (user_id);
