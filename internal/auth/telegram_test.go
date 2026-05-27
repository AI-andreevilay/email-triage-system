package auth

import "testing"

func TestVerifyTelegramLogin(t *testing.T) {
	login := TelegramLogin{
		ID:        42,
		Username:  "alice",
		FirstName: "Alice",
		AuthDate:  1710000000,
		Hash:      "f595c4510eb9f26cb2338beef8f17c2aabee529358a6bc2256c5399c2bbe7cf1",
	}

	if err := VerifyTelegramLogin(login, "123456:ABC-DEF"); err != nil {
		t.Fatalf("expected valid login: %v", err)
	}

	login.Hash = "bad"
	if err := VerifyTelegramLogin(login, "123456:ABC-DEF"); err == nil {
		t.Fatal("expected invalid login")
	}
}
