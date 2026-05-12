package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

type LabelWorker struct {
	store  *storage.Postgres
	broker *broker.RabbitMQ
}

func NewLabelWorker(store *storage.Postgres, messageBroker *broker.RabbitMQ) (*LabelWorker, error) {
	if store == nil || messageBroker == nil {
		return nil, errors.New("label worker dependencies are not configured")
	}
	return &LabelWorker{
		store:  store,
		broker: messageBroker,
	}, nil
}

func (w *LabelWorker) Run(ctx context.Context) error {
	deliveries, err := w.broker.ConsumeClassifiedEmails(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("classified email consumer channel closed")
			}
			if err := w.handleDelivery(ctx, delivery); err != nil {
				return err
			}
		}
	}
}

func (w *LabelWorker) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	err := w.processClassifiedEmail(ctx, delivery.Body)
	if err == nil {
		return delivery.Ack(false)
	}

	if errors.Is(err, errDropDelivery) {
		return delivery.Ack(false)
	}

	if nackErr := delivery.Nack(false, true); nackErr != nil {
		return fmt.Errorf("process delivery: %w; nack: %w", err, nackErr)
	}
	return nil
}

func (w *LabelWorker) processClassifiedEmail(ctx context.Context, body []byte) error {
	var event broker.ClassifiedEmailEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return errDropDelivery
	}

	if event.Mode != "apply" {
		return nil
	}

	if err := mockApplyLabel(event); err != nil {
		return err
	}

	err := w.store.MarkEmailLabelApplied(
		ctx,
		event.UserID,
		event.Classification.GmailMessageID,
		event.Classification.PredictedLabel,
	)
	if err != nil {
		if errors.Is(err, storage.ErrEmailMessageNotFound) {
			return errDropDelivery
		}
		return err
	}

	return nil
}

func mockApplyLabel(_ broker.ClassifiedEmailEvent) error {
	return nil
}
