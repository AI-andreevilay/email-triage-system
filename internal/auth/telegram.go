package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidTelegramLogin = errors.New("invalid telegram login")

type TelegramLogin struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
	AuthDate  int64  `json:"auth_date"`
	Hash      string `json:"hash"`
}

func VerifyTelegramLogin(login TelegramLogin, botToken string) error {
	if botToken == "" {
		return errors.New("telegram bot token is not configured")
	}
	if login.ID == 0 || login.AuthDate == 0 || login.Hash == "" {
		return ErrInvalidTelegramLogin
	}

	values := url.Values{}
	values.Set("id", strconv.FormatInt(login.ID, 10))
	values.Set("auth_date", strconv.FormatInt(login.AuthDate, 10))
	if login.Username != "" {
		values.Set("username", login.Username)
	}
	if login.FirstName != "" {
		values.Set("first_name", login.FirstName)
	}
	if login.LastName != "" {
		values.Set("last_name", login.LastName)
	}
	if login.PhotoURL != "" {
		values.Set("photo_url", login.PhotoURL)
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values.Get(key))
	}
	dataCheckString := strings.Join(parts, "\n")

	secret := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secret[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expectedHash), []byte(strings.ToLower(login.Hash))) {
		return fmt.Errorf("%w: hash mismatch", ErrInvalidTelegramLogin)
	}
	return nil
}

func (l TelegramLogin) DisplayName() string {
	switch {
	case l.FirstName != "" && l.LastName != "":
		return l.FirstName + " " + l.LastName
	case l.FirstName != "":
		return l.FirstName
	case l.LastName != "":
		return l.LastName
	default:
		return l.Username
	}
}
