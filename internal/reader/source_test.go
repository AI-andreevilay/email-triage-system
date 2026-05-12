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
