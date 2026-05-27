package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

func (p *Postgres) GetUserByTelegramID(ctx context.Context, telegramID int64) (models.User, error) {
	return p.getUser(ctx, `WHERE telegram_id = $1`, telegramID)
}

func (p *Postgres) GetUserByID(ctx context.Context, id string) (models.User, error) {
	return p.getUser(ctx, `WHERE id = $1`, id)
}

func (p *Postgres) getUser(ctx context.Context, where string, arg any) (models.User, error) {
	var user models.User
	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, telegram_id, telegram_username, display_name, role, enabled, created_at, updated_at
		 FROM users `+where,
		arg,
	).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.DisplayName,
		&user.Role,
		&user.Enabled,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}
