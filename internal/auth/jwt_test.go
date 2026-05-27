package auth

import (
	"testing"
	"time"
)

func TestTokenManagerIssueVerify(t *testing.T) {
	manager := NewTokenManager(Config{
		JWTSecret:   "secret",
		JWTIssuer:   "email-triage-system",
		JWTAudience: []string{"email-triage-system", "pg-ops-console"},
		TokenTTL:    time.Hour,
	})
	manager.now = func() time.Time { return time.Unix(1000, 0) }

	token, err := manager.Issue(Principal{
		UserID:   "f030bdb4-31a5-45e5-a73f-53f194ab2499",
		Role:     "admin",
		Provider: "telegram",
	})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	principal, err := manager.Verify(token, "pg-ops-console")
	if err != nil {
		t.Fatalf("verify token: %v", err)
	}
	if principal.UserID != "f030bdb4-31a5-45e5-a73f-53f194ab2499" || principal.Role != "admin" || principal.Provider != "telegram" {
		t.Fatalf("unexpected principal: %+v", principal)
	}
}

func TestTokenManagerVerifyRejectsWrongAudienceAndExpiredToken(t *testing.T) {
	manager := NewTokenManager(Config{
		JWTSecret:   "secret",
		JWTIssuer:   "email-triage-system",
		JWTAudience: []string{"email-triage-system"},
		TokenTTL:    time.Hour,
	})
	manager.now = func() time.Time { return time.Unix(1000, 0) }

	token, err := manager.Issue(Principal{UserID: "user-id", Role: "user", Provider: "telegram"})
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	if _, err := manager.Verify(token, "pg-ops-console"); err == nil {
		t.Fatal("expected wrong audience to be rejected")
	}

	manager.now = func() time.Time { return time.Unix(5000, 0) }
	if _, err := manager.Verify(token, "email-triage-system"); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}
