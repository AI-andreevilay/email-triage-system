package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/bzelijah/email-triage-system/internal/classifier"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

type Handler struct {
	store      *storage.Postgres
	reader     *reader.MockReader
	classifier *classifier.Classifier
}

func NewRouter(store *storage.Postgres, mockReader *reader.MockReader, messageClassifier *classifier.Classifier) (http.Handler, error) {
	if store == nil || mockReader == nil || messageClassifier == nil {
		return nil, errors.New("api dependencies are not configured")
	}

	h := &Handler{
		store:      store,
		reader:     mockReader,
		classifier: messageClassifier,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthz)
	mux.HandleFunc("POST /scans", h.createScan)
	return mux, nil
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
