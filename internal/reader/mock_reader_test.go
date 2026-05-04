package reader

import (
	"context"
	"testing"
)

func TestMockReaderListMessages(t *testing.T) {
	reader := NewMockReader()

	messages, err := reader.ListMessages(context.Background(), "user_1")
	if err != nil {
		t.Fatalf("ListMessages returned error: %v", err)
	}
	if len(messages) == 0 {
		t.Fatal("expected non-empty messages")
	}
	if messages[0].ID == "" {
		t.Fatal("expected message id")
	}
}
