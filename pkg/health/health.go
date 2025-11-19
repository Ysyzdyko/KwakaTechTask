package health

import (
	"context"

	"menu-parser/internal/usecase"
	"menu-parser/pkg/database"
	"menu-parser/pkg/queue"
)

// HealthService implements usecase.HealthCheckService
type HealthService struct {
	db    *database.MongoDB
	queue *queue.RabbitMQ
}

// NewHealthService creates a new health service
func NewHealthService(db *database.MongoDB, queue *queue.RabbitMQ) usecase.HealthCheckService {
	return &HealthService{
		db:    db,
		queue: queue,
	}
}

func (s *HealthService) CheckDatabase(ctx context.Context) error {
	return s.db.HealthCheck(ctx)
}

func (s *HealthService) CheckQueue() error {
	return s.queue.HealthCheck()
}



