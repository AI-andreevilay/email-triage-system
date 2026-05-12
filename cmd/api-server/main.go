package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bzelijah/email-triage-system/internal/api"
	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/config"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

func main() {
	cfg := config.Load()

	pg, err := storage.NewPostgres(context.Background(), cfg.PostgresURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pg.Close()

	mq, err := broker.NewRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer mq.Close()

	mockReader := reader.NewMockReader()
	router, err := api.NewRouter(pg, mockReader, mq)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
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
