package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

type Handler struct {
	store  *storage.Postgres
	reader *reader.Source
	broker *broker.RabbitMQ
	mux    *http.ServeMux
}

func NewHandler(store *storage.Postgres, emailReader *reader.Source, messageBroker *broker.RabbitMQ) (*Handler, error) {
	if store == nil || emailReader == nil || messageBroker == nil {
		return nil, errors.New("api dependencies are not configured")
	}

	h := &Handler{
		store:  store,
		reader: emailReader,
		broker: messageBroker,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /scans", h.createScan)
	mux.HandleFunc("GET /scans/{id}", h.getScan)
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
