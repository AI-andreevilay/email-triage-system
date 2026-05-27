package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPPort    string
	PostgresURL string
	RabbitMQURL string
	EmailSource string

	GmailCredentialsFile   string
	GmailTokenFile         string
	GmailUserID            string
	GmailReadMaxResults    int64
	GmailReadQuery         string
	LabelWorkerConcurrency int
	ScheduledScanInterval  time.Duration
	ScheduledScanUserID    string
	ScheduledScanMode      string
	ScheduledScanQuery     string
	ScheduledScanMarkRead  bool

	JWTSecret        string
	JWTIssuer        string
	JWTAudience      []string
	TelegramBotToken string
	AuthTokenTTL     time.Duration
}

func Load() Config {
	port := getEnv("HTTP_PORT", "8080")
	postgresURL := getEnv("POSTGRES_URL", "postgres://postgres:postgres@localhost:5432/email_triage?sslmode=disable")
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	emailSource := getEnv("EMAIL_SOURCE", "mock")
	gmailCredentialsFile := getEnv("GMAIL_CREDENTIALS_FILE", "secrets/gmail_credentials.json")
	gmailTokenFile := getEnv("GMAIL_TOKEN_FILE", "secrets/gmail_token.json")
	gmailUserID := getEnv("GMAIL_USER_ID", "me")
	gmailReadMaxResults := getEnvInt64("GMAIL_READ_MAX_RESULTS", 100)
	gmailReadQuery := getEnv("GMAIL_READ_QUERY", "in:inbox -in:trash")
	labelWorkerConcurrency := getEnvInt("LABEL_WORKER_CONCURRENCY", 4)
	scheduledScanInterval := getEnvDuration("SCHEDULED_SCAN_INTERVAL", 0)
	scheduledScanUserID := getEnv("SCHEDULED_SCAN_USER_ID", "")
	scheduledScanMode := getEnv("SCHEDULED_SCAN_MODE", "dry_run")
	scheduledScanQuery := getEnv("SCHEDULED_SCAN_QUERY", "")
	scheduledScanMarkRead := getEnvBool("SCHEDULED_SCAN_MARK_READ", false)
	jwtSecret := getEnv("JWT_SECRET", "")
	jwtIssuer := getEnv("JWT_ISSUER", "email-triage-system")
	jwtAudience := getEnvStringSlice("JWT_AUDIENCE", []string{"email-triage-system", "pg-ops-console"})
	telegramBotToken := getEnv("TELEGRAM_BOT_TOKEN", "")
	authTokenTTL := getEnvDuration("AUTH_TOKEN_TTL", 24*time.Hour)
	return Config{
		HTTPPort:               port,
		PostgresURL:            postgresURL,
		RabbitMQURL:            rabbitMQURL,
		EmailSource:            emailSource,
		GmailCredentialsFile:   gmailCredentialsFile,
		GmailTokenFile:         gmailTokenFile,
		GmailUserID:            gmailUserID,
		GmailReadMaxResults:    gmailReadMaxResults,
		GmailReadQuery:         gmailReadQuery,
		LabelWorkerConcurrency: labelWorkerConcurrency,
		ScheduledScanInterval:  scheduledScanInterval,
		ScheduledScanUserID:    scheduledScanUserID,
		ScheduledScanMode:      scheduledScanMode,
		ScheduledScanQuery:     scheduledScanQuery,
		ScheduledScanMarkRead:  scheduledScanMarkRead,
		JWTSecret:              jwtSecret,
		JWTIssuer:              jwtIssuer,
		JWTAudience:            jwtAudience,
		TelegramBotToken:       telegramBotToken,
		AuthTokenTTL:           authTokenTTL,
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

func getEnvInt(key string, fallback int) int {
	parsed := getEnvInt64(key, int64(fallback))
	if parsed > int64(^uint(0)>>1) {
		return fallback
	}
	return int(parsed)
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(v)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvStringSlice(key string, fallback []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
