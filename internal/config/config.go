package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPPort      string
	PostgresURL   string
	RabbitMQURL   string
	MigrationsDir string
	EmailSource   string

	GmailCredentialsFile string
	GmailTokenFile       string
	GmailUserID          string
	GmailReadMaxResults  int64
	GmailReadQuery       string
}

func Load() Config {
	port := getEnv("HTTP_PORT", "8080")
	postgresURL := getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/email_triage?sslmode=disable")
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	migrationsDir := getEnv("MIGRATIONS_DIR", "migrations")
	emailSource := getEnv("EMAIL_SOURCE", "mock")
	gmailCredentialsFile := getEnv("GMAIL_CREDENTIALS_FILE", "secrets/gmail_credentials.json")
	gmailTokenFile := getEnv("GMAIL_TOKEN_FILE", "secrets/gmail_token.json")
	gmailUserID := getEnv("GMAIL_USER_ID", "me")
	gmailReadMaxResults := getEnvInt64("GMAIL_READ_MAX_RESULTS", 100)
	gmailReadQuery := getEnv("GMAIL_READ_QUERY", "in:inbox -in:trash")
	return Config{
		HTTPPort:             port,
		PostgresURL:          postgresURL,
		RabbitMQURL:          rabbitMQURL,
		MigrationsDir:        migrationsDir,
		EmailSource:          emailSource,
		GmailCredentialsFile: gmailCredentialsFile,
		GmailTokenFile:       gmailTokenFile,
		GmailUserID:          gmailUserID,
		GmailReadMaxResults:  gmailReadMaxResults,
		GmailReadQuery:       gmailReadQuery,
	}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	if parsed <= 0 {
		return fallback
	}
	return parsed
}
