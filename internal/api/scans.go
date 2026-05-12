package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bzelijah/email-triage-system/internal/broker"
	storagemodels "github.com/bzelijah/email-triage-system/internal/storage/models"
)

const (
	scanModeDryRun = "dry_run"
	scanModeApply  = "apply"
	defaultUserID  = "user_1"
)

type createScanRequest struct {
	Mode string `json:"mode"`
}

type createScanResponse struct {
	RunID          int64  `json:"run_id"`
	UserID         string `json:"user_id"`
	Mode           string `json:"mode"`
	Status         string `json:"status"`
	TotalFound     int    `json:"total_found"`
	TotalProcessed int    `json:"total_processed"`
	TotalFailed    int    `json:"total_failed"`
}

func (h *Handler) createScan(w http.ResponseWriter, r *http.Request) {
	var req createScanRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	mode := req.Mode
	if mode == "" {
		mode = scanModeDryRun
	}
	if mode != scanModeDryRun && mode != scanModeApply {
		writeJSONError(w, http.StatusBadRequest, "mode must be dry_run or apply")
		return
	}

	runID, err := h.store.CreateScanRun(r.Context(), storagemodels.ScanRun{
		UserID: defaultUserID,
		Mode:   mode,
		Status: "running",
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create scan run")
		return
	}

	failRun := func() {
		_ = h.store.CompleteScanRun(r.Context(), storagemodels.ScanRun{
			ID:             runID,
			Status:         "failed",
			TotalFound:     0,
			TotalProcessed: 0,
			TotalFailed:    0,
		})
	}

	messages, err := h.reader.ListMessages(r.Context(), defaultUserID)
	if err != nil {
		failRun()
		writeJSONError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	totalProcessed := 0
	totalFailed := 0

	for _, message := range messages {
		err := h.broker.PublishRawEmail(r.Context(), broker.RawEmailEvent{
			ScanRunID:   runID,
			UserID:      defaultUserID,
			Mode:        mode,
			PublishedAt: time.Now().UTC(),
			Message: broker.RawEmailMessage{
				GmailMessageID: message.ID,
				ThreadID:       message.ThreadID,
				From:           message.From,
				Subject:        message.Subject,
				BodySnippet:    message.BodySnippet,
			},
		})
		if err != nil {
			totalFailed++
			continue
		}
		totalProcessed++
	}

	runStatus := "queued"
	if totalFailed > 0 {
		runStatus = "queued_with_errors"
	}

	if err := h.store.CompleteScanRun(r.Context(), storagemodels.ScanRun{
		ID:             runID,
		Status:         runStatus,
		TotalFound:     len(messages),
		TotalProcessed: totalProcessed,
		TotalFailed:    totalFailed,
	}); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to complete scan run")
		return
	}

	response := createScanResponse{
		RunID:          runID,
		UserID:         defaultUserID,
		Mode:           mode,
		Status:         runStatus,
		TotalFound:     len(messages),
		TotalProcessed: totalProcessed,
		TotalFailed:    totalFailed,
	}
	writeJSON(w, http.StatusCreated, response)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
