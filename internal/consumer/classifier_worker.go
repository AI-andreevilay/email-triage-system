package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/classifier"
	"github.com/bzelijah/email-triage-system/internal/reader"
	"github.com/bzelijah/email-triage-system/internal/rules"
	"github.com/bzelijah/email-triage-system/internal/storage"
	storagemodels "github.com/bzelijah/email-triage-system/internal/storage/models"
)

var errDropDelivery = errors.New("drop delivery")

type ClassifierWorker struct {
	store      *storage.Postgres
	broker     *broker.RabbitMQ
	classifier *classifier.Classifier
}

func NewClassifierWorker(store *storage.Postgres, messageBroker *broker.RabbitMQ, messageClassifier *classifier.Classifier) (*ClassifierWorker, error) {
	if store == nil || messageBroker == nil || messageClassifier == nil {
		return nil, errors.New("classifier worker dependencies are not configured")
	}
	return &ClassifierWorker{
		store:      store,
		broker:     messageBroker,
		classifier: messageClassifier,
	}, nil
}

func (w *ClassifierWorker) Run(ctx context.Context) error {
	deliveries, err := w.broker.ConsumeRawEmails(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("raw email consumer channel closed")
			}
			if err := w.handleDelivery(ctx, delivery); err != nil {
				return err
			}
		}
	}
}

func (w *ClassifierWorker) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	err := w.processRawEmail(ctx, delivery.Body)
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

func (w *ClassifierWorker) processRawEmail(ctx context.Context, body []byte) error {
	var event broker.RawEmailEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return errDropDelivery
	}

	userRules, err := w.store.ListEnabledUserRules(ctx, event.UserID)
	if err != nil {
		return err
	}

	result := w.classifier.Classify(reader.Message{
		ID:          event.Message.GmailMessageID,
		ThreadID:    event.Message.ThreadID,
		From:        event.Message.From,
		Subject:     event.Message.Subject,
		BodySnippet: event.Message.BodySnippet,
	}, toClassificationRules(userRules))

	status := "classified"
	if event.Mode == "dry_run" {
		status = "dry_run"
	}

	now := time.Now().UTC()
	err = w.store.InsertEmailMessage(ctx, storagemodels.EmailMessage{
		UserID:         event.UserID,
		GmailMessageID: event.Message.GmailMessageID,
		PredictedLabel: result.Label,
		AppliedLabel:   nil,
		Confidence:     result.Confidence,
		Reason:         result.Reason,
		Status:         status,
		ProcessedAt:    &now,
	})
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyProcessed) {
			return nil
		}
		return err
	}

	if event.Mode == "apply" {
		if err := w.broker.PublishClassifiedEmail(ctx, broker.ClassifiedEmailEvent{
			ScanRunID:    event.ScanRunID,
			UserID:       event.UserID,
			Mode:         event.Mode,
			ClassifiedAt: time.Now().UTC(),
			Classification: broker.ClassifiedEmailMessage{
				GmailMessageID: event.Message.GmailMessageID,
				PredictedLabel: result.Label,
				Confidence:     result.Confidence,
			},
		}); err != nil {
			return err
		}
	}

	return nil
}

func toClassificationRules(in []storagemodels.UserRule) []rules.Rule {
	result := make([]rules.Rule, 0, len(in))
	for _, rule := range in {
		result = append(result, rules.Rule{
			RuleType:    rule.RuleType,
			Operator:    rule.Operator,
			RuleValue:   rule.RuleValue,
			TargetLabel: rule.TargetLabel,
			Enabled:     rule.Enabled,
			Priority:    rule.Priority,
		})
	}
	return result
}
