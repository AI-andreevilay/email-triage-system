package models

import "time"

type ScanRun struct {
	ID             int64
	UserID         string
	Mode           string
	Status         string
	StartedAt      time.Time
	FinishedAt     *time.Time
	TotalFound     int
	TotalProcessed int
	TotalFailed    int
}
