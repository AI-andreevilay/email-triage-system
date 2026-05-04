package config

import "os"

type Config struct {
	HTTPPort      string
	PostgresURL   string
	MigrationsDir string
}

func Load() Config {
	port := getEnv("HTTP_PORT", "8080")
	postgresURL := getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/email_triage?sslmode=disable")
	migrationsDir := getEnv("MIGRATIONS_DIR", "migrations")
	return Config{
		HTTPPort:      port,
		PostgresURL:   postgresURL,
		MigrationsDir: migrationsDir,
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
