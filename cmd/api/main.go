package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"menu-parser/internal/repository"
	httpDelivery "menu-parser/internal/transport/http"
	"menu-parser/internal/usecase"
	"menu-parser/pkg/config"
	"menu-parser/pkg/database"
	"menu-parser/pkg/health"
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

	rabbitmq, err := rabbitmqQueue.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}
	defer rabbitmq.Close()

	menuRepo := repository.NewMenuRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	sheetsParser, err := parser.NewSheetsParser(cfg.GoogleSheetsCredentialsPath)
	if err != nil {
		log.Fatalf("Failed to initialize parser: %v", err)
	}

	queuePublisher := rabbitmqQueue.NewQueuePublisher(rabbitmq)

	healthService := health.NewHealthService(db, rabbitmq)

	menuUseCase := usecase.NewMenuUseCase(menuRepo, taskRepo, sheetsParser, queuePublisher)
	productUseCase := usecase.NewProductUseCase(menuRepo, auditRepo, queuePublisher)
	healthUseCase := usecase.NewHealthUseCase(healthService)

	router := httpDelivery.SetupRouter(menuUseCase, productUseCase, healthUseCase)

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.APIHost, cfg.APIPort),
		Handler: router,
	}

	go func() {
		log.Printf("Starting API server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
