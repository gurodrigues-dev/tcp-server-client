package mq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gurodrigues-dev/tcp-server-client/internal/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQProducer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *messaging.RabbitMQOpts
}

func NewRabbitMQProducer(opts *messaging.RabbitMQOpts) (*RabbitMQProducer, error) {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s%s",
		opts.Username,
		opts.Password,
		opts.Host,
		opts.Port,
		opts.VHost,
	)
	log.Printf("[RabbitMQProducer] Connecting to RabbitMQ %s", dsn)

	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	producer := &RabbitMQProducer{
		connection: conn,
		channel:    ch,
		config:     opts,
	}

	if err := producer.declareExchange(opts.ExchangeName, opts.ExchangeType); err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	if err := producer.validateConnection(); err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to validate connection: %w", err)
	}

	log.Println("[RabbitMQProducer] Connection established successfully")

	return producer, nil
}

func (p *RabbitMQProducer) Publish(name, topic string, message interface{}) error {
	if p.channel == nil || p.channel.IsClosed() {
		return fmt.Errorf("channel is not open")
	}

	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	log.Printf("[RabbitMQProducer] Message to publish on topic %s: %s", topic, string(body))

	props := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	}

	err = p.channel.Publish(
		name,
		topic,
		false,
		false,
		props,
	)

	if err != nil {
		log.Printf("[RabbitMQProducer] Failed to publish message on topic %s: %v", topic, err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("[RabbitMQProducer] Message published on topic %s", topic)

	return nil
}

func (p *RabbitMQProducer) Close() error {
	log.Println("[RabbitMQProducer] Closing connection")

	var errs []error

	if p.channel != nil && !p.channel.IsClosed() {
		if err := p.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if p.connection != nil && !p.connection.IsClosed() {
		if err := p.connection.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to close producer: %v", errs)
	}

	log.Println("[RabbitMQProducer] Connection closed successfully")
	return nil
}

func (p *RabbitMQProducer) declareExchange(exchangeName, exchangeType string) error {
	err := p.channel.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange %s: %w", exchangeName, err)
	}

	log.Printf("[RabbitMQProducer] Exchange %s,type %s declared successfully", exchangeName, exchangeType)

	return nil
}

func (p *RabbitMQProducer) validateConnection() error {
	if p.connection == nil || p.connection.IsClosed() {
		return fmt.Errorf("connection is closed")
	}

	if p.channel == nil || p.channel.IsClosed() {
		return fmt.Errorf("channel is closed")
	}

	log.Println("[RabbitMQProducer] Connection validated successfully")

	return nil
}
