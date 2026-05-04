package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/bzelijah/email-triage-system/internal/config"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal(err)
	}

	if err := ensureMigrationsTable(ctx, db); err != nil {
		log.Fatal(err)
	}

	paths, err := migrationFiles(cfg.MigrationsDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, path := range paths {
		applied, err := isApplied(ctx, db, filepath.Base(path))
		if err != nil {
			log.Fatal(err)
		}
		if applied {
			continue
		}
		if err := applyFile(ctx, db, path); err != nil {
			log.Fatal(err)
		}
		log.Printf("applied migration %s", filepath.Base(path))
	}
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			paths = append(paths, filepath.Join(dir, name))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

func isApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(
		ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`,
		version,
	).Scan(&exists)
	return exists, err
}

func applyFile(ctx context.Context, db *sql.DB, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("execute migration %s: %w", path, err)
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO schema_migrations (version) VALUES ($1)`,
		filepath.Base(path),
	); err != nil {
		return fmt.Errorf("save migration %s: %w", path, err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
