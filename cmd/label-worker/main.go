package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/config"
	"github.com/bzelijah/email-triage-system/internal/consumer"
	"github.com/bzelijah/email-triage-system/internal/gmail"
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

	gmailClient, err := gmail.NewClient(
		context.Background(),
		cfg.GmailCredentialsFile,
		cfg.GmailTokenFile,
		cfg.GmailUserID,
	)
	if err != nil {
		log.Fatal(fmt.Errorf("init gmail client: %w", err))
	}

	worker, err := consumer.NewLabelWorker(pg, mq, gmailClient, cfg.LabelWorkerConcurrency)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("label-worker started with concurrency=%d", cfg.LabelWorkerConcurrency)
	if err := worker.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
