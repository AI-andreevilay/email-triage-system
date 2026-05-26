package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bzelijah/email-triage-system/internal/broker"
	"github.com/bzelijah/email-triage-system/internal/gmail"
	"github.com/bzelijah/email-triage-system/internal/storage"
)

type LabelWorker struct {
	store       *storage.Postgres
	broker      *broker.RabbitMQ
	gmailClient *gmail.Client
	concurrency int
	ackMu       sync.Mutex
}

func NewLabelWorker(store *storage.Postgres, messageBroker *broker.RabbitMQ, gmailClient *gmail.Client, concurrency int) (*LabelWorker, error) {
	if store == nil || messageBroker == nil || gmailClient == nil {
		return nil, errors.New("label worker dependencies are not configured")
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	return &LabelWorker{
		store:       store,
		broker:      messageBroker,
		gmailClient: gmailClient,
		concurrency: concurrency,
	}, nil
}

func (w *LabelWorker) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	deliveries, err := w.broker.ConsumeClassifiedEmails(runCtx)
	if err != nil {
		return err
	}

	errCh := make(chan error, w.concurrency)
	var wg sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-runCtx.Done():
					return
				case delivery, ok := <-deliveries:
					if !ok {
						return
					}
					if err := w.handleDelivery(runCtx, delivery); err != nil {
						select {
						case errCh <- err:
						default:
						}
						cancel()
						return
					}
				}
			}
		}()
	}

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()

	for {
		select {
		case <-ctx.Done():
			cancel()
			<-doneCh
			return nil
		case err := <-errCh:
			cancel()
			<-doneCh
			return err
		case <-doneCh:
			return nil
		}
	}
}

func (w *LabelWorker) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	err := w.processClassifiedEmail(ctx, delivery.Body)
	if err == nil {
		w.ackMu.Lock()
		defer w.ackMu.Unlock()
		return delivery.Ack(false)
	}

	if errors.Is(err, errDropDelivery) {
		w.ackMu.Lock()
		defer w.ackMu.Unlock()
		return delivery.Ack(false)
	}

	w.ackMu.Lock()
	defer w.ackMu.Unlock()
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

	labelID, err := w.gmailClient.EnsureLabel(ctx, event.Classification.PredictedLabel)
	if err != nil {
		if gmail.IsPermanentError(err) {
			return errDropDelivery
		}
		return err
	}

	if err := w.gmailClient.ApplyLabelToMessage(ctx, event.Classification.GmailMessageID, labelID); err != nil {
		if gmail.IsPermanentError(err) {
			return errDropDelivery
		}
		return err
	}

	err = w.store.MarkEmailLabelApplied(
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
