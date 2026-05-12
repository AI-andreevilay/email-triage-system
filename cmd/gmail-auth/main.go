package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bzelijah/email-triage-system/internal/config"
	"github.com/bzelijah/email-triage-system/internal/gmail"
)

func main() {
	cfg := config.Load()

	oauthCfg, err := gmail.LoadOAuthConfig(cfg.GmailCredentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	const callbackURL = "http://localhost:8090/oauth2/callback"
	oauthCfg.RedirectURL = callbackURL

	state, err := randomState()
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}

		select {
		case codeCh <- code:
		default:
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Authorization received. You can close this tab."))
	})

	server := &http.Server{
		Addr:              ":8090",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if listenErr := server.ListenAndServe(); listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			errCh <- listenErr
		}
	}()

	authURL := gmail.BuildAuthURL(oauthCfg, state)
	fmt.Println("Open this URL in browser and authorize Gmail access:")
	fmt.Println(authURL)

	var code string
	select {
	case <-ctx.Done():
		_ = server.Shutdown(context.Background())
		log.Fatal("authorization cancelled")
	case listenErr := <-errCh:
		log.Fatal(listenErr)
	case code = <-codeCh:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	case <-time.After(3 * time.Minute):
		_ = server.Shutdown(context.Background())
		log.Fatal("authorization timeout")
	}

	token, err := gmail.ExchangeCode(context.Background(), oauthCfg, code)
	if err != nil {
		log.Fatal(err)
	}

	if err := gmail.SaveToken(cfg.GmailTokenFile, token); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Saved token to %s\n", cfg.GmailTokenFile)
}

func randomState() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
