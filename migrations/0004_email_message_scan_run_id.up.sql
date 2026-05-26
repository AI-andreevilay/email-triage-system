ALTER TABLE email_messages
    ADD COLUMN IF NOT EXISTS scan_run_id BIGINT REFERENCES scan_runs(id);

CREATE INDEX IF NOT EXISTS email_messages_scan_run_id_idx
    ON email_messages (scan_run_id);
