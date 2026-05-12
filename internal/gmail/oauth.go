package gmail

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmailv1 "google.golang.org/api/gmail/v1"
)

func LoadOAuthConfig(credentialsFile string) (*oauth2.Config, error) {
	content, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}
	return google.ConfigFromJSON(content, gmailv1.GmailModifyScope)
}

func BuildAuthURL(cfg *oauth2.Config, state string) string {
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func ExchangeCode(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	return cfg.Exchange(ctx, code)
}

func SaveToken(tokenFile string, token *oauth2.Token) error {
	if token == nil {
		return errors.New("token is nil")
	}

	dir := filepath.Dir(tokenFile)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(token)
}

func loadToken(tokenFile string) (*oauth2.Token, error) {
	file, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var token oauth2.Token
	if err := json.NewDecoder(file).Decode(&token); err != nil {
		return nil, err
	}

	return &token, nil
}

func AuthenticatedHTTPClient(ctx context.Context, credentialsFile, tokenFile string) (*http.Client, error) {
	cfg, err := LoadOAuthConfig(credentialsFile)
	if err != nil {
		return nil, err
	}

	token, err := loadToken(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("load token file %s: %w", tokenFile, err)
	}

	return cfg.Client(ctx, token), nil
}
