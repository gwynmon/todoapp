package broker

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const Exchange = "task-events"

type Event struct {
	Type      string         `json:"event_type"`
	TaskID    int            `json:"task_id"`
	UserID    int            `json:"user_id"`
	Timestamp time.Time      `json:"timestamp"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type Producer struct {
	ch *amqp.Channel
}

func NewProducer(conn *amqp.Connection) (*Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		Exchange,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Producer{ch: ch}, nil
}

func (p *Producer) Publish(ctx context.Context, routingKey string, event Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.ch.PublishWithContext(ctx,
		Exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (p *Producer) Close() error {
	return p.ch.Close()
}
