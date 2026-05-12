package broker

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewRabbitMQ(url string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(EmailEventsExchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(EmailRawQueue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	if err := ch.QueueBind(EmailRawQueue, EmailRawRoutingKey, EmailEventsExchange, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(EmailClassifiedQueue, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	if err := ch.QueueBind(EmailClassifiedQueue, EmailClassifiedKey, EmailEventsExchange, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &RabbitMQ{
		conn: conn,
		ch:   ch,
	}, nil
}

func (r *RabbitMQ) PublishRawEmail(ctx context.Context, event RawEmailEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.ch.PublishWithContext(
		ctx,
		EmailEventsExchange,
		EmailRawRoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (r *RabbitMQ) ConsumeRawEmails(ctx context.Context) (<-chan amqp.Delivery, error) {
	if err := r.ch.Qos(10, 0, false); err != nil {
		return nil, err
	}

	return r.ch.ConsumeWithContext(
		ctx,
		EmailRawQueue,
		"classifier-worker",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) PublishClassifiedEmail(ctx context.Context, event ClassifiedEmailEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return r.ch.PublishWithContext(
		ctx,
		EmailEventsExchange,
		EmailClassifiedKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (r *RabbitMQ) ConsumeClassifiedEmails(ctx context.Context) (<-chan amqp.Delivery, error) {
	if err := r.ch.Qos(10, 0, false); err != nil {
		return nil, err
	}

	return r.ch.ConsumeWithContext(
		ctx,
		EmailClassifiedQueue,
		"label-worker",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (r *RabbitMQ) Close() error {
	if r == nil {
		return nil
	}

	var firstErr error

	if r.ch != nil {
		if err := r.ch.Close(); err != nil {
			firstErr = err
		}
	}

	if r.conn != nil {
		if err := r.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}
