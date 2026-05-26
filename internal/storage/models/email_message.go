package models

import "time"

type EmailMessage struct {
	ScanRunID      *int64
	UserID         string
	GmailMessageID string
	PredictedLabel string
	AppliedLabel   *string
	Confidence     float64
	Reason         string
	Status         string
	ProcessedAt    *time.Time
}
