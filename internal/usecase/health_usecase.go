package usecase

import (
	"context"
	"time"
)

// HealthCheckService defines the interface for health checks
type HealthCheckService interface {
	CheckDatabase(ctx context.Context) error
	CheckQueue() error
}

// HealthUseCase handles health check logic
type HealthUseCase struct {
	healthService HealthCheckService
}

// NewHealthUseCase creates a new HealthUseCase
func NewHealthUseCase(healthService HealthCheckService) *HealthUseCase {
	return &HealthUseCase{
		healthService: healthService,
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

// Check performs health check for all services
func (uc *HealthUseCase) Check(ctx context.Context) *HealthResponse {
	services := make(map[string]string)

	// Check database
	if err := uc.healthService.CheckDatabase(ctx); err != nil {
		services["database"] = "error"
	} else {
		services["database"] = "ok"
	}

	// Check queue
	if err := uc.healthService.CheckQueue(); err != nil {
		services["queue"] = "error"
	} else {
		services["queue"] = "ok"
	}

	status := "healthy"
	for _, s := range services {
		if s == "error" {
			status = "unhealthy"
			break
		}
	}

	return &HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Services:  services,
	}
}

