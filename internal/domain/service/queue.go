package service

import (
	"menu-parser/internal/domain/entity"
)

type QueuePublisher interface {
	PublishMenuParsingTask(taskID string) error
	PublishProductStatusEvent(event *entity.ProductStatusChangeEvent) error
}

type Message struct {
	Body        []byte
	DeliveryTag uint64
}

type QueueConsumer interface {
	ConsumeMenuParsingTasks() (<-chan Message, error)
	ConsumeProductStatusEvents() (<-chan Message, error)
	AckMessage(deliveryTag uint64) error
	NackMessage(deliveryTag uint64, requeue bool) error
}
