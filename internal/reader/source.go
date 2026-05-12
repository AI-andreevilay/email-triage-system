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
	result := make([]Message, 0)
	err := s.IterateMessages(ctx, userID, func(batch []Message) error {
		result = append(result, batch...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Source) IterateMessages(ctx context.Context, userID string, visit func(batch []Message) error) error {
	if visit == nil {
		return errors.New("visit callback is required")
	}

	if s.source == SourceMock {
		messages, err := s.mock.ListMessages(ctx, userID)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			return nil
		}
		return visit(messages)
	}

	pageToken := ""
	for {
		items, nextPageToken, err := s.gmail.ListMessagesPage(ctx, s.maxResults, s.query, pageToken)
		if err != nil {
			return err
		}

		batch := toMessages(items)
		if len(batch) > 0 {
			if err := visit(batch); err != nil {
				return err
			}
		}

		if nextPageToken == "" || nextPageToken == pageToken {
			return nil
		}
		pageToken = nextPageToken
	}
}

func toMessages(items []gmail.Message) []Message {
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
	return result
}
