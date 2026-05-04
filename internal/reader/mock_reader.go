package reader

import "context"

type MockReader struct{}

func NewMockReader() *MockReader {
	return &MockReader{}
}

func (r *MockReader) ListMessages(ctx context.Context, userID string) ([]Message, error) {
	return []Message{
		{
			ID:          "msg_1001",
			ThreadID:    "thr_9001",
			From:        "jobs@company.com",
			Subject:     "Backend Go Engineer opportunity",
			BodySnippet: "We reviewed your profile and would like to schedule an interview.",
		},
		{
			ID:          "msg_1002",
			ThreadID:    "thr_9002",
			From:        "alerts@bank.com",
			Subject:     "Your card transaction was approved",
			BodySnippet: "Transaction amount $42.15 at Coffee Shop.",
		},
		{
			ID:          "msg_1003",
			ThreadID:    "thr_9003",
			From:        "security@google.com",
			Subject:     "New sign-in detected",
			BodySnippet: "We noticed a new sign-in to your account.",
		},
		{
			ID:          "msg_1004",
			ThreadID:    "thr_9004",
			From:        "newsletter@store.com",
			Subject:     "Weekend sale 50% off",
			BodySnippet: "Promo code inside. Limited offer.",
		},
		{
			ID:          "msg_1005",
			ThreadID:    "thr_9005",
			From:        "friend@social.app",
			Subject:     "You have new notifications",
			BodySnippet: "5 new likes and 2 comments.",
		},
	}, nil
}
