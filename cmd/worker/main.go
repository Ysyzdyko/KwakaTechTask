package main

import (
	"context"
	"log"

	"menu-parser/internal/repository"
	"menu-parser/internal/transport/queue"
	"menu-parser/internal/usecase"
	"menu-parser/pkg/config"
	"menu-parser/pkg/database"
	"menu-parser/pkg/parser"
	rabbitmqQueue "menu-parser/pkg/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewMongoDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close(context.Background())

	rabbitmqInstance, err := rabbitmqQueue.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}
	defer rabbitmqInstance.Close()

	menuRepo := repository.NewMenuRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	sheetsParser, err := parser.NewSheetsParser(cfg.GoogleSheetsCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize parser: %v", err)
	}

	queuePublisher := rabbitmqQueue.NewQueuePublisher(rabbitmqInstance)
	queueConsumer, err := rabbitmqQueue.NewQueueConsumer(rabbitmqInstance)
	if err != nil {
		log.Fatalf("Failed to initialize queue consumer: %v", err)
	}

	menuUseCase := usecase.NewMenuUseCase(menuRepo, taskRepo, sheetsParser, queuePublisher)
	productUseCase := usecase.NewProductUseCase(menuRepo, auditRepo, queuePublisher)

	consumer := queue.NewConsumer(menuUseCase, productUseCase, taskRepo, queueConsumer)

	consumer.Start()
}
