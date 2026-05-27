package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bzelijah/email-triage-system/internal/auth"
	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

type Handler struct {
	store      *storage.Postgres
	reader     *reader.Source
	broker     *broker.RabbitMQ
	authConfig auth.Config
	tokens     *auth.TokenManager
	mux        *http.ServeMux
}

func NewHandler(store *storage.Postgres, emailReader *reader.Source, messageBroker *broker.RabbitMQ, authConfig auth.Config) (*Handler, error) {
	if store == nil || emailReader == nil || messageBroker == nil {
		return nil, errors.New("api dependencies are not configured")
	}
	authConfig = authConfig.Normalized()

	h := &Handler{
		store:      store,
		reader:     emailReader,
		broker:     messageBroker,
		authConfig: authConfig,
		tokens:     auth.NewTokenManager(authConfig),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /auth/telegram", h.telegramAuth)
	mux.HandleFunc("GET /auth/me", h.requireAuth(h.me))
	mux.HandleFunc("POST /scans", h.requireAuth(h.createScan))
	mux.HandleFunc("GET /scans/{id}", h.requireAuth(h.getScan))
	h.mux = mux
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return auth.Middleware(h.tokens, h.authConfig.JWTIssuer, next)
}
