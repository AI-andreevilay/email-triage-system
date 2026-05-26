package storage

import (
	"context"
	"database/sql"
	"errors"

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

func (p *Postgres) UpdateScanRunProgress(ctx context.Context, run models.ScanRun) error {
	_, err := p.db.ExecContext(
		ctx,
		`UPDATE scan_runs
		 SET total_found = $1, total_processed = $2, total_failed = $3
		 WHERE id = $4`,
		run.TotalFound,
		run.TotalProcessed,
		run.TotalFailed,
		run.ID,
	)
	return err
}

func (p *Postgres) GetScanRun(ctx context.Context, id int64) (models.ScanRun, error) {
	var run models.ScanRun
	err := p.db.QueryRowContext(
		ctx,
		`SELECT id, user_id, mode, status, started_at, finished_at, total_found, total_processed, total_failed
		 FROM scan_runs
		 WHERE id = $1`,
		id,
	).Scan(
		&run.ID,
		&run.UserID,
		&run.Mode,
		&run.Status,
		&run.StartedAt,
		&run.FinishedAt,
		&run.TotalFound,
		&run.TotalProcessed,
		&run.TotalFailed,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ScanRun{}, ErrScanRunNotFound
		}
		return models.ScanRun{}, err
	}
	return run, nil
}
