package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/storage"
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
	Result string `json:"result"`
	RunID  int64  `json:"run_id"`
	UserID string `json:"user_id"`
	Mode   string `json:"mode"`
	Status string `json:"status"`
}

type scanResponse struct {
	RunID          int64        `json:"run_id"`
	UserID         string       `json:"user_id"`
	Mode           string       `json:"mode"`
	Status         string       `json:"status"`
	StartedAt      time.Time    `json:"started_at"`
	FinishedAt     *time.Time   `json:"finished_at,omitempty"`
	TotalFound     int          `json:"total_found"`
	TotalProcessed int          `json:"total_processed"`
	TotalFailed    int          `json:"total_failed"`
	EmailStatus    statusCounts `json:"email_status"`
}

type statusCounts struct {
	DryRun     int `json:"dry_run"`
	Classified int `json:"classified"`
	Applied    int `json:"applied"`
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
		Status: "enqueuing",
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to create scan run")
		return
	}

	go h.enqueueScan(context.Background(), runID, defaultUserID, mode)

	writeJSON(w, http.StatusAccepted, createScanResponse{
		Result: "ok",
		RunID:  runID,
		UserID: defaultUserID,
		Mode:   mode,
		Status: "enqueuing",
	})
}

func (h *Handler) enqueueScan(ctx context.Context, runID int64, userID, mode string) {
	failRun := func(totalFound, totalProcessed, totalFailed int) {
		_ = h.store.CompleteScanRun(ctx, storagemodels.ScanRun{
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

	if err := h.reader.IterateMessages(ctx, userID, func(batch []reader.Message) error {
		totalFound += len(batch)

		for _, message := range batch {
			err := h.broker.PublishRawEmail(ctx, broker.RawEmailEvent{
				ScanRunID:   runID,
				UserID:      userID,
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

		return h.store.UpdateScanRunProgress(ctx, storagemodels.ScanRun{
			ID:             runID,
			TotalFound:     totalFound,
			TotalProcessed: totalProcessed,
			TotalFailed:    totalFailed,
		})
	}); err != nil {
		log.Printf("failed to enqueue scan run_id=%d user_id=%s error=%v", runID, userID, err)
		failRun(totalFound, totalProcessed, totalFailed)
		return
	}

	runStatus := "queued"
	if totalFailed > 0 {
		runStatus = "queued_with_errors"
	}

	if err := h.store.CompleteScanRun(ctx, storagemodels.ScanRun{
		ID:             runID,
		Status:         runStatus,
		TotalFound:     totalFound,
		TotalProcessed: totalProcessed,
		TotalFailed:    totalFailed,
	}); err != nil {
		log.Printf("failed to complete scan run_id=%d user_id=%s error=%v", runID, userID, err)
		return
	}

}

func (h *Handler) getScan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSONError(w, http.StatusBadRequest, "invalid scan id")
		return
	}

	run, err := h.store.GetScanRun(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrScanRunNotFound) {
			writeJSONError(w, http.StatusNotFound, "scan run not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "failed to get scan run")
		return
	}

	counts, err := h.store.CountEmailMessagesByScanRun(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to get scan run counters")
		return
	}

	writeJSON(w, http.StatusOK, scanResponse{
		RunID:          run.ID,
		UserID:         run.UserID,
		Mode:           run.Mode,
		Status:         run.Status,
		StartedAt:      run.StartedAt,
		FinishedAt:     run.FinishedAt,
		TotalFound:     run.TotalFound,
		TotalProcessed: run.TotalProcessed,
		TotalFailed:    run.TotalFailed,
		EmailStatus: statusCounts{
			DryRun:     counts["dry_run"],
			Classified: counts["classified"],
			Applied:    counts["applied"],
		},
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
