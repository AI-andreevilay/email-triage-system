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

CREATE TEMP TABLE user_id_map (
    old_user_id TEXT PRIMARY KEY,
    new_user_id UUID NOT NULL
) ON COMMIT DROP;

INSERT INTO user_id_map (old_user_id, new_user_id)
SELECT DISTINCT
    user_id,
    CASE
        WHEN user_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
            THEN user_id::UUID
        ELSE gen_random_uuid()
    END
FROM (
    SELECT user_id FROM scan_runs WHERE user_id IS NOT NULL
    UNION
    SELECT user_id FROM email_messages WHERE user_id IS NOT NULL
    UNION
    SELECT user_id FROM user_rules WHERE user_id IS NOT NULL
) legacy_ids;

INSERT INTO users (id, telegram_id, display_name, role, enabled)
SELECT new_user_id, -ROW_NUMBER() OVER (ORDER BY old_user_id), old_user_id, 'user', FALSE
FROM user_id_map
ON CONFLICT (id) DO NOTHING;

ALTER TABLE scan_runs
    ALTER COLUMN user_id TYPE UUID USING (
        CASE
            WHEN user_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
                THEN user_id::UUID
            ELSE (SELECT new_user_id FROM user_id_map WHERE old_user_id = scan_runs.user_id)
        END
    );

ALTER TABLE email_messages
    ALTER COLUMN user_id TYPE UUID USING (
        CASE
            WHEN user_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
                THEN user_id::UUID
            ELSE (SELECT new_user_id FROM user_id_map WHERE old_user_id = email_messages.user_id)
        END
    );

DROP INDEX IF EXISTS user_rules_unique_rule;

ALTER TABLE user_rules
    ALTER COLUMN user_id TYPE UUID USING (
        CASE
            WHEN user_id IS NULL THEN NULL
            WHEN user_id ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
                THEN user_id::UUID
            ELSE (SELECT new_user_id FROM user_id_map WHERE old_user_id = user_rules.user_id)
        END
    );

CREATE UNIQUE INDEX IF NOT EXISTS user_rules_unique_rule
    ON user_rules (user_id, rule_type, operator, rule_value, target_label)
    NULLS NOT DISTINCT;

CREATE INDEX IF NOT EXISTS scan_runs_user_id_idx
    ON scan_runs (user_id);

CREATE INDEX IF NOT EXISTS email_messages_user_id_idx
    ON email_messages (user_id);
