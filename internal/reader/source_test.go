package reader

import (
	"context"
	"testing"
)

func TestSourceListMessagesMock(t *testing.T) {
	src, err := NewSource(SourceMock, NewMockReader(), nil, 10, "")
	if err != nil {
		t.Fatalf("NewSource returned error: %v", err)
	}

	messages, err := src.ListMessages(context.Background(), "user_1")
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected non-empty messages")
	}
}

func TestSourceIterateMessagesMock(t *testing.T) {
	src, err := NewSource(SourceMock, NewMockReader(), nil, 10, "")
	if err != nil {
		t.Fatalf("NewSource returned error: %v", err)
	}

	total := 0
	err = src.IterateMessages(context.Background(), "user_1", func(batch []Message) error {
		total += len(batch)
		return nil
	})
	if err != nil {
		t.Fatalf("IterateMessages returned error: %v", err)
	}
	if total == 0 {
		t.Fatal("expected non-empty batches")
	}
}
