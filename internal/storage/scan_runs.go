package storage

import (
	"context"

	"github.com/bzelijah/email-triage-system/internal/storage/models"
)

func (p *Postgres) CreateScanRun(ctx context.Context, run models.ScanRun) (int64, error) {
	var id int64
	err := p.db.QueryRowContext(
		ctx,
		`INSERT INTO scan_runs (user_id, mode, status)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		run.UserID,
		run.Mode,
		run.Status,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (p *Postgres) CompleteScanRun(ctx context.Context, run models.ScanRun) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE scan_runs
		 SET status = $1, finished_at = NOW(), total_found = $2, total_processed = $3, total_failed = $4
		 WHERE id = $5`,
		run.Status,
		run.TotalFound,
		run.TotalProcessed,
		run.TotalFailed,
		run.ID,
	)
	return err
}
