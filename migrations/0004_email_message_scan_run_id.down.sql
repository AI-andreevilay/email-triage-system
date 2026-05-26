DROP INDEX IF EXISTS email_messages_scan_run_id_idx;

ALTER TABLE email_messages
    DROP COLUMN IF EXISTS scan_run_id;
