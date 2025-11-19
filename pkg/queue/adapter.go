package queue

import (
	"fmt"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/service"

	"github.com/streadway/amqp"
)

type QueuePublisherAdapter struct {
	rabbitmq *RabbitMQ
}

func NewQueuePublisher(rabbitmq *RabbitMQ) service.QueuePublisher {
	return &QueuePublisherAdapter{rabbitmq: rabbitmq}
}

func (q *QueuePublisherAdapter) PublishMenuParsingTask(taskID string) error {
	return q.rabbitmq.PublishMenuParsingTask(taskID)
}

func (q *QueuePublisherAdapter) PublishProductStatusEvent(event *entity.ProductStatusChangeEvent) error {
	return q.rabbitmq.PublishProductStatusEvent(event)
}

type QueueConsumerAdapter struct {
	rabbitmq          *RabbitMQ
	menuParsingMsgs   <-chan amqp.Delivery
	productStatusMsgs <-chan amqp.Delivery
	menuOutput        chan service.Message
	productOutput     chan service.Message
	deliveryMap       map[uint64]amqp.Delivery
}

func NewQueueConsumer(rabbitmq *RabbitMQ) (service.QueueConsumer, error) {
	menuMsgs, err := rabbitmq.ConsumeMenuParsingTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to consume menu parsing tasks: %w", err)
	}

	productMsgs, err := rabbitmq.ConsumeProductStatusEvents()
	if err != nil {
		return nil, fmt.Errorf("failed to consume product status events: %w", err)
	}

	adapter := &QueueConsumerAdapter{
		rabbitmq:          rabbitmq,
		menuParsingMsgs:   menuMsgs,
		productStatusMsgs: productMsgs,
		menuOutput:        make(chan service.Message, 100),
		productOutput:     make(chan service.Message, 100),
		deliveryMap:       make(map[uint64]amqp.Delivery),
	}

	go func() {
		for msg := range menuMsgs {
			adapter.deliveryMap[msg.DeliveryTag] = msg
			adapter.menuOutput <- service.Message{
				Body:       msg.Body,
				DeliveryTag: msg.DeliveryTag,
			}
		}
		close(adapter.menuOutput)
	}()

	go func() {
		for msg := range productMsgs {
			adapter.deliveryMap[msg.DeliveryTag] = msg
			adapter.productOutput <- service.Message{
				Body:       msg.Body,
				DeliveryTag: msg.DeliveryTag,
			}
		}
		close(adapter.productOutput)
	}()

	return adapter, nil
}

func (q *QueueConsumerAdapter) ConsumeMenuParsingTasks() (<-chan service.Message, error) {
	return q.menuOutput, nil
}

func (q *QueueConsumerAdapter) ConsumeProductStatusEvents() (<-chan service.Message, error) {
	return q.productOutput, nil
}

func (q *QueueConsumerAdapter) AckMessage(deliveryTag uint64) error {
	delivery, exists := q.deliveryMap[deliveryTag]
	if !exists {
		return fmt.Errorf("delivery tag %d not found", deliveryTag)
	}
	delete(q.deliveryMap, deliveryTag)
	return delivery.Ack(false)
}

func (q *QueueConsumerAdapter) NackMessage(deliveryTag uint64, requeue bool) error {
	delivery, exists := q.deliveryMap[deliveryTag]
	if !exists {
		return fmt.Errorf("delivery tag %d not found", deliveryTag)
	}
	delete(q.deliveryMap, deliveryTag)
	return delivery.Nack(false, requeue)
}
