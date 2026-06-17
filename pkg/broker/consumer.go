package broker

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	channel *amqp.Channel
	queue   string
}

func NewConsumer(conn *amqp.Connection) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	q, err := ch.QueueDeclare(
		"notifier-queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, err
	}

	err = ch.QueueBind(
		q.Name,
		"task.*",
		"task-events",
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		return nil, err
	}

	return &Consumer{
		channel: ch,
		queue:   q.Name,
	}, nil
}

func (c *Consumer) Consume() (<-chan amqp.Delivery, error) {
	return c.channel.Consume(
		c.queue,
		"",
		true, // auto ack
		false,
		false,
		false,
		nil,
	)
}

func (c *Consumer) Close() error {
	return c.channel.Close()
}
