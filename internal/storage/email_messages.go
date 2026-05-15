package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
	"github.com/jackc/pgx/v5/pgconn"
)

func (p *Postgres) InsertEmailMessage(ctx context.Context, message models.EmailMessage) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO email_messages
		(user_id, gmail_message_id, predicted_label, applied_label, confidence, reason, status, processed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		message.UserID,
		message.GmailMessageID,
		message.PredictedLabel,
		message.AppliedLabel,
		message.Confidence,
		message.Reason,
		message.Status,
		message.ProcessedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyProcessed
		}
		return err
	}
	return nil
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
		`SELECT user_id, gmail_message_id, predicted_label, applied_label, confidence, reason, status, processed_at
		 FROM email_messages
		 WHERE user_id = $1 AND gmail_message_id = $2`,
		userID,
		gmailMessageID,
	).Scan(
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
