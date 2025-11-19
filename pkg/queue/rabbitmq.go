package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"menu-parser/pkg/config"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	conn               *amqp.Connection
	channel            *amqp.Channel
	menuParsingQueue   string
	productStatusQueue string
	dlqQueue           string
}

func NewRabbitMQ(cfg *config.Config) (*RabbitMQ, error) {
	conn, err := amqp.Dial(cfg.RabbitMQURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	dlqName := cfg.RabbitMQDLQQueue
	_, err = ch.QueueDeclare(
		dlqName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	menuParsingQueue := cfg.RabbitMQMenuParsingQueue
	_, err = ch.QueueDeclare(
		menuParsingQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlqName,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare menu-parsing queue: %w", err)
	}

	productStatusQueue := cfg.RabbitMQProductStatusQueue
	_, err = ch.QueueDeclare(
		productStatusQueue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": dlqName,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare product-status queue: %w", err)
	}

	return &RabbitMQ{
		conn:               conn,
		channel:            ch,
		menuParsingQueue:   menuParsingQueue,
		productStatusQueue: productStatusQueue,
		dlqQueue:           dlqName,
	}, nil
}

func (r *RabbitMQ) PublishMenuParsingTask(taskID string) error {
	message := map[string]string{
		"task_id": taskID,
	}
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.channel.Publish(
		"",
		r.menuParsingQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (r *RabbitMQ) PublishProductStatusEvent(event interface{}) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = r.channel.Publish(
		"",
		r.productStatusQueue,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (r *RabbitMQ) ConsumeMenuParsingTasks() (<-chan amqp.Delivery, error) {
	err := r.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := r.channel.Consume(
		r.menuParsingQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return msgs, nil
}

func (r *RabbitMQ) ConsumeProductStatusEvents() (<-chan amqp.Delivery, error) {
	err := r.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := r.channel.Consume(
		r.productStatusQueue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	return msgs, nil
}

func (r *RabbitMQ) Close() error {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) HealthCheck() error {
	if r.conn.IsClosed() {
		return fmt.Errorf("RabbitMQ connection is closed")
	}
	return nil
}
