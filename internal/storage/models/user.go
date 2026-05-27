package models

import "time"

type User struct {
	ID               string
	TelegramID       int64
	TelegramUsername *string
	DisplayName      *string
	Role             string
	Enabled          bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
