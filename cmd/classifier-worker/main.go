package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/classifier"
	"github.com/bzelijah/email-triage-system/internal/config"
	"github.com/bzelijah/email-triage-system/internal/consumer"
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

	worker, err := consumer.NewClassifierWorker(pg, mq, classifier.New())
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("classifier-worker started")
	if err := worker.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
