package rmqrpc

import (
	"context"
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Config -.
type Config struct {
	URL      string
	WaitTime time.Duration
	Attempts int
}

// Connection -.
type Connection struct {
	ConsumerExchange string
	Config
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Delivery   <-chan amqp.Delivery
}

// New -.
func New(consumerExchange string, cfg Config) *Connection {
	conn := &Connection{
		ConsumerExchange: consumerExchange,
		Config:           cfg,
	}

	return conn
}

// AttemptConnect -.
func (c *Connection) AttemptConnect() error {
	return c.AttemptConnectContext(context.Background())
}

// AttemptConnectContext retries Dial until success, attempts exhausted, or ctx canceled (F01).
func (c *Connection) AttemptConnectContext(ctx context.Context) error {
	var err error

	for i := c.Attempts; i > 0; i-- {
		if ctx.Err() != nil {
			return fmt.Errorf("rmq_rpc - AttemptConnectContext: %w", ctx.Err())
		}

		if err = c.connect(); err == nil {
			return nil
		}

		log.Printf("RabbitMQ is trying to connect, attempts left: %d", i)

		select {
		case <-ctx.Done():
			return fmt.Errorf("rmq_rpc - AttemptConnectContext: %w", ctx.Err())
		case <-time.After(c.WaitTime):
		}
	}

	if err != nil {
		return fmt.Errorf("rmq_rpc - AttemptConnect - c.connect: %w", err)
	}

	return nil
}

func (c *Connection) connect() error {
	var err error

	c.Connection, err = amqp.Dial(c.URL)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	c.Channel, err = c.Connection.Channel()
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	err = c.Channel.ExchangeDeclare(
		c.ConsumerExchange,
		"fanout",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Connection.Channel: %w", err)
	}

	queue, err := c.Channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueDeclare: %w", err)
	}

	err = c.Channel.QueueBind(
		queue.Name,
		"",
		c.ConsumerExchange,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.QueueBind: %w", err)
	}

	c.Delivery, err = c.Channel.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("c.Channel.Consume: %w", err)
	}

	return nil
}
