package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bzelijah/email-triage-system/internal/rules"
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
	RunID            int64  `json:"run_id"`
	UserID           string `json:"user_id"`
	Mode             string `json:"mode"`
	Status           string `json:"status"`
	TotalFound       int    `json:"total_found"`
	TotalProcessed   int    `json:"total_processed"`
	TotalFailed      int    `json:"total_failed"`
	AlreadyProcessed int    `json:"already_processed"`
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

	userRules, err := h.store.ListEnabledUserRules(r.Context(), defaultUserID)
	if err != nil {
		failRun()
		writeJSONError(w, http.StatusInternalServerError, "failed to load user rules")
		return
	}
	classificationRules := toClassificationRules(userRules)

	totalProcessed := 0
	totalFailed := 0
	alreadyProcessed := 0

	for _, message := range messages {
		classification := h.classifier.Classify(message, classificationRules)

		var appliedLabel *string
		messageStatus := "classified"
		if mode == scanModeDryRun {
			messageStatus = "dry_run"
		}
		if mode == scanModeApply {
			applied := classification.Label
			appliedLabel = &applied
			messageStatus = "applied"
		}

		now := time.Now().UTC()
		err := h.store.InsertEmailMessage(r.Context(), storagemodels.EmailMessage{
			UserID:         defaultUserID,
			GmailMessageID: message.ID,
			PredictedLabel: classification.Label,
			AppliedLabel:   appliedLabel,
			Confidence:     classification.Confidence,
			Status:         messageStatus,
			ProcessedAt:    &now,
		})
		if err != nil {
			if errors.Is(err, storage.ErrAlreadyProcessed) {
				alreadyProcessed++
				continue
			}
			totalFailed++
			continue
		}
		totalProcessed++
	}

	runStatus := "completed"
	if totalFailed > 0 {
		runStatus = "completed_with_errors"
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
		RunID:            runID,
		UserID:           defaultUserID,
		Mode:             mode,
		Status:           runStatus,
		TotalFound:       len(messages),
		TotalProcessed:   totalProcessed,
		TotalFailed:      totalFailed,
		AlreadyProcessed: alreadyProcessed,
	}
	writeJSON(w, http.StatusCreated, response)
}

func toClassificationRules(in []storagemodels.UserRule) []rules.Rule {
	result := make([]rules.Rule, 0, len(in))
	for _, rule := range in {
		result = append(result, rules.Rule{
			RuleType:    rule.RuleType,
			RuleValue:   rule.RuleValue,
			TargetLabel: rule.TargetLabel,
			Enabled:     rule.Enabled,
			Priority:    rule.Priority,
		})
	}
	return result
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}
