package queue

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/repository"
	"menu-parser/internal/domain/service"
	"menu-parser/internal/usecase"
)

const maxRetries = 3

type Consumer struct {
	menuUseCase    *usecase.MenuUseCase
	productUseCase *usecase.ProductUseCase
	taskRepo       repository.TaskRepository
	queueConsumer  service.QueueConsumer
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewConsumer(
	menuUseCase *usecase.MenuUseCase,
	productUseCase *usecase.ProductUseCase,
	taskRepo repository.TaskRepository,
	queueConsumer service.QueueConsumer,
) *Consumer {
	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		menuUseCase:    menuUseCase,
		productUseCase: productUseCase,
		taskRepo:       taskRepo,
		queueConsumer:  queueConsumer,
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (c *Consumer) Start() {
	log.Println("Queue consumer started")

	go c.processMenuParsingTasks()

	go c.processProductStatusEvents()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down queue consumer...")
	c.Shutdown()
}

func (c *Consumer) processMenuParsingTasks() {
	msgs, err := c.queueConsumer.ConsumeMenuParsingTasks()
	if err != nil {
		log.Printf("Error consuming menu parsing tasks: %v", err)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.handleMenuParsingTask(msg)
		}
	}
}

func (c *Consumer) handleMenuParsingTask(msg service.Message) {
	var message map[string]string
	if err := json.Unmarshal(msg.Body, &message); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		c.queueConsumer.NackMessage(msg.DeliveryTag, false)
		return
	}

	taskID := message["task_id"]
	if taskID == "" {
		log.Printf("Empty task_id in message")
		c.queueConsumer.NackMessage(msg.DeliveryTag, false)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get task
	task, err := c.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		log.Printf("Error getting task %s: %v", taskID, err)
		c.queueConsumer.NackMessage(msg.DeliveryTag, true) // Requeue
		return
	}

	// Check retry count
	if task.RetryCount >= maxRetries {
		log.Printf("Task %s exceeded max retries", taskID)
		c.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusFailed, nil, "Max retries exceeded")
		c.queueConsumer.NackMessage(msg.DeliveryTag, false) // Don't requeue, goes to DLQ
		return
	}

	// Process menu parsing
	err = c.menuUseCase.ProcessMenuParsing(ctx, taskID)
	if err != nil {
		log.Printf("Error processing menu parsing: %v", err)

		// Increment retry count
		c.taskRepo.IncrementRetryCount(ctx, taskID)

		// Calculate exponential backoff delay
		delay := time.Duration(math.Pow(2, float64(task.RetryCount))) * time.Second
		time.Sleep(delay)

		// Update status and requeue
		c.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusQueued, nil, err.Error())
		c.queueConsumer.NackMessage(msg.DeliveryTag, true) // Requeue for retry
		return
	}

	// ACK message after successful processing
	if err := c.queueConsumer.AckMessage(msg.DeliveryTag); err != nil {
		log.Printf("Error ACKing message: %v", err)
	}
	log.Printf("Successfully processed task %s", taskID)
}

func (c *Consumer) processProductStatusEvents() {
	msgs, err := c.queueConsumer.ConsumeProductStatusEvents()
	if err != nil {
		log.Printf("Error consuming product status events: %v", err)
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.handleProductStatusEvent(msg)
		}
	}
}

func (c *Consumer) handleProductStatusEvent(msg service.Message) {
	var event entity.ProductStatusChangeEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		c.queueConsumer.NackMessage(msg.DeliveryTag, false)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.productUseCase.ProcessProductStatusEvent(ctx, &event); err != nil {
		log.Printf("Error processing product status event: %v", err)
		c.queueConsumer.NackMessage(msg.DeliveryTag, true) // Requeue
		return
	}

	// ACK message after successful processing
	if err := c.queueConsumer.AckMessage(msg.DeliveryTag); err != nil {
		log.Printf("Error ACKing message: %v", err)
	}
	log.Printf("Processed product status event for product %s: %s -> %s",
		event.ProductID, event.OldStatus, event.NewStatus)
}

func (c *Consumer) Shutdown() {
	c.cancel()

	// Give workers time to finish current tasks
	time.Sleep(5 * time.Second)

	log.Println("Queue consumer stopped")
}
