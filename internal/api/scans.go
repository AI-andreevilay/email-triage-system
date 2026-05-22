package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/reader"
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

	failRun := func(totalFound, totalProcessed, totalFailed int) {
		_ = h.store.CompleteScanRun(r.Context(), storagemodels.ScanRun{
			ID:             runID,
			Status:         "failed",
			TotalFound:     totalFound,
			TotalProcessed: totalProcessed,
			TotalFailed:    totalFailed,
		})
	}

	totalFound := 0
	totalProcessed := 0
	totalFailed := 0

	if err := h.reader.IterateMessages(r.Context(), defaultUserID, func(batch []reader.Message) error {
		totalFound += len(batch)

		for _, message := range batch {
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

		return nil
	}); err != nil {
		log.Printf("failed to list messages for scan run_id=%d user_id=%s source error: %v", runID, defaultUserID, err)
		failRun(totalFound, totalProcessed, totalFailed)
		writeJSONError(w, http.StatusInternalServerError, "failed to list messages")
		return
	}

	runStatus := "queued"
	if totalFailed > 0 {
		runStatus = "queued_with_errors"
	}

	if err := h.store.CompleteScanRun(r.Context(), storagemodels.ScanRun{
		ID:             runID,
		Status:         runStatus,
		TotalFound:     totalFound,
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
		TotalFound:     totalFound,
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
