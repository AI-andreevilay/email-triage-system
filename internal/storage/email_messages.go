package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

func (p *Postgres) InsertEmailMessage(ctx context.Context, message models.EmailMessage) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO email_messages
		(scan_run_id, user_id, gmail_message_id, predicted_label, applied_label, confidence, reason, status, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		message.ScanRunID,
		message.UserID,
		message.GmailMessageID,
		message.PredictedLabel,
		message.AppliedLabel,
		message.Confidence,
		message.Reason,
		message.Status,
		message.ProcessedAt,
	)
	return err
}

func (p *Postgres) UpsertEmailMessage(ctx context.Context, message models.EmailMessage) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO email_messages
		(scan_run_id, user_id, gmail_message_id, predicted_label, applied_label, confidence, reason, status, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (user_id, gmail_message_id)
		DO UPDATE SET
			scan_run_id = EXCLUDED.scan_run_id,
			predicted_label = EXCLUDED.predicted_label,
			applied_label = EXCLUDED.applied_label,
			confidence = EXCLUDED.confidence,
			reason = EXCLUDED.reason,
			status = EXCLUDED.status,
			processed_at = EXCLUDED.processed_at`,
		message.ScanRunID,
		message.UserID,
		message.GmailMessageID,
		message.PredictedLabel,
		message.AppliedLabel,
		message.Confidence,
		message.Reason,
		message.Status,
		message.ProcessedAt,
	)
	return err
}

func (p *Postgres) MarkEmailLabelApplied(ctx context.Context, userID, gmailMessageID, appliedLabel string) error {
	result, err := p.db.ExecContext(
		ctx,
		`UPDATE email_messages
		 SET applied_label = $1, status = $2, processed_at = NOW()
		 WHERE user_id = $3 AND gmail_message_id = $4`,
		appliedLabel,
		"applied",
		userID,
		gmailMessageID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEmailMessageNotFound
	}

	return nil
}

func (p *Postgres) GetEmailMessage(ctx context.Context, userID, gmailMessageID string) (models.EmailMessage, error) {
	var message models.EmailMessage
	err := p.db.QueryRowContext(
		ctx,
		`SELECT scan_run_id, user_id, gmail_message_id, predicted_label, applied_label, confidence, reason, status, processed_at
		 FROM email_messages
		 WHERE user_id = $1 AND gmail_message_id = $2`,
		userID,
		gmailMessageID,
	).Scan(
		&message.ScanRunID,
		&message.UserID,
		&message.GmailMessageID,
		&message.PredictedLabel,
		&message.AppliedLabel,
		&message.Confidence,
		&message.Reason,
		&message.Status,
		&message.ProcessedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.EmailMessage{}, ErrEmailMessageNotFound
		}
		return models.EmailMessage{}, err
	}
	return message, nil
}

func (p *Postgres) CountEmailMessagesByScanRun(ctx context.Context, scanRunID int64) (map[string]int, error) {
	rows, err := p.db.QueryContext(
		ctx,
		`SELECT status, COUNT(*)
		 FROM email_messages
		 WHERE scan_run_id = $1
		 GROUP BY status`,
		scanRunID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
