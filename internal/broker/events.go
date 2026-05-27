package broker

import "time"

const (
	EmailEventsExchange  = "email.events"
	EmailRawRoutingKey   = "email.raw"
	EmailRawQueue        = "email.raw"
	EmailClassifiedKey   = "email.classified"
	EmailClassifiedQueue = "email.classified"
)

type RawEmailEvent struct {
	ScanRunID   int64           `json:"scan_run_id"`
	UserID      string          `json:"user_id"`
	Mode        string          `json:"mode"`
	MarkRead    bool            `json:"mark_read"`
	PublishedAt time.Time       `json:"published_at"`
	Message     RawEmailMessage `json:"message"`
}

type RawEmailMessage struct {
	GmailMessageID string `json:"gmail_message_id"`
	ThreadID       string `json:"thread_id"`
	From           string `json:"from"`
	Subject        string `json:"subject"`
	BodySnippet    string `json:"body_snippet"`
}

type ClassifiedEmailEvent struct {
	ScanRunID      int64                  `json:"scan_run_id"`
	UserID         string                 `json:"user_id"`
	Mode           string                 `json:"mode"`
	MarkRead       bool                   `json:"mark_read"`
	ClassifiedAt   time.Time              `json:"classified_at"`
	Classification ClassifiedEmailMessage `json:"classification"`
}

type ClassifiedEmailMessage struct {
	GmailMessageID string  `json:"gmail_message_id"`
	PredictedLabel string  `json:"predicted_label"`
	Confidence     float64 `json:"confidence"`
}
