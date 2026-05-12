package storage

import (
	"context"
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
