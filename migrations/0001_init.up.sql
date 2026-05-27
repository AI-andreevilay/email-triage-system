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

CREATE TABLE IF NOT EXISTS email_messages (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    gmail_message_id TEXT NOT NULL,
    predicted_label TEXT NOT NULL,
    applied_label TEXT,
    confidence DOUBLE PRECISION NOT NULL,
    reason TEXT,
    status TEXT NOT NULL,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, gmail_message_id)
);

CREATE TABLE IF NOT EXISTS scan_runs (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    mode TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ,
    total_found INTEGER NOT NULL DEFAULT 0,
    total_processed INTEGER NOT NULL DEFAULT 0,
    total_failed INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS user_rules (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    rule_type TEXT NOT NULL,
    operator TEXT NOT NULL,
    rule_value TEXT NOT NULL,
    target_label TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
