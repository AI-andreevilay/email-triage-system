package reader

import (
	"context"
	"errors"

	"github.com/bzelijah/email-triage-system/internal/gmail"
)

const (
	SourceMock  = "mock"
	SourceGmail = "gmail"
)

type Source struct {
	source     string
	mock       *MockReader
	gmail      *gmail.Client
	maxResults int64
	query      string
}

func NewSource(source string, mock *MockReader, gmailClient *gmail.Client, maxResults int64, query string) (*Source, error) {
	if source != SourceMock && source != SourceGmail {
		return nil, errors.New("unsupported email source")
	}
	if source == SourceMock && mock == nil {
		return nil, errors.New("mock reader is required")
	}
	if source == SourceGmail && gmailClient == nil {
		return nil, errors.New("gmail client is required")
	}
	if maxResults <= 0 {
		maxResults = 20
	}

	return &Source{
		source:     source,
		mock:       mock,
		gmail:      gmailClient,
		maxResults: maxResults,
		query:      query,
	}, nil
}

func (s *Source) ListMessages(ctx context.Context, userID string) ([]Message, error) {
	if s.source == SourceMock {
		return s.mock.ListMessages(ctx, userID)
	}

	items, err := s.gmail.ListMessages(ctx, s.maxResults, s.query)
	if err != nil {
		return nil, err
	}

	result := make([]Message, 0, len(items))
	for _, item := range items {
		result = append(result, Message{
			ID:          item.ID,
			ThreadID:    item.ThreadID,
			From:        item.From,
			Subject:     item.Subject,
			BodySnippet: item.Snippet,
		})
	}
	return result, nil
}
