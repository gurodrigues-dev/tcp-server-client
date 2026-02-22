package mq

import (
	"fmt"
	"log"

	"github.com/gurodrigues-dev/tcp-server-client/internal/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *messaging.RabbitMQOpts
	queue      string
	handler    func([]byte) error
	done       chan struct{}
}

func NewRabbitMQConsumer(opts *messaging.RabbitMQOpts) (*RabbitMQConsumer, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.VHost,
	)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	err = ch.ExchangeDeclare(
		opts.ExchangeName,
		opts.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	q, err := ch.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = ch.QueueBind(
		q.Name,
		"#",
		opts.ExchangeName,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &RabbitMQConsumer{
		connection: conn,
		channel:    ch,
		config:     opts,
		queue:      q.Name,
		done:       make(chan struct{}),
	}, nil
}

func (c *RabbitMQConsumer) Subscribe(topic string, handler func([]byte) error) error {
	c.handler = handler
	return nil
}

func (c *RabbitMQConsumer) Start() error {
	if c.handler == nil {
		return fmt.Errorf("handler not set, call Subscribe first")
	}

	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for {
			select {
			case <-c.done:
				log.Printf("[RabbitMQConsumer] Receive signal received")
				return
			case msg, ok := <-msgs:
				if !ok {
					log.Printf("[RabbitMQConsumer] Message channel closed")
					return
				}

				log.Printf("[RabbitMQConsumer] Message received:")
				log.Printf("  Body (string): %s", string(msg.Body))
				log.Printf("  Body (len): %d bytes", len(msg.Body))

				if err := c.handler(msg.Body); err != nil {
					log.Printf("[RabbitMQConsumer] Error in handler: %v", err)
					msg.Nack(false, true)
				} else {
					msg.Ack(false)
				}
			}
		}
	}()

	log.Printf("[RabbitMQConsumer] Consumer started and waiting for messages")
	return nil
}

func (c *RabbitMQConsumer) Stop() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}

	err := c.channel.Cancel("", false)
	if err != nil {
		return fmt.Errorf("failed to cancel consumer: %w", err)
	}

	return nil
}

func (c *RabbitMQConsumer) Close() error {
	var errs []error

	if c.channel != nil && !c.channel.IsClosed() {
		if err := c.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if c.connection != nil && !c.connection.IsClosed() {
		if err := c.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing: %v", errs)
	}

	return nil
}
