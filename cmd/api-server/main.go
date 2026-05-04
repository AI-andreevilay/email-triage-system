package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bzelijah/email-triage-system/internal/api"
	"github.com/bzelijah/email-triage-system/internal/config"
)

func main() {
	cfg := config.Load()
	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("api-server listening on :%s", cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
