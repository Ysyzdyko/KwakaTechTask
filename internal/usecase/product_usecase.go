package usecase

import (
	"context"
	"fmt"
	"time"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/repository"
	"menu-parser/internal/domain/service"
)

// ProductUseCase handles product-related business logic
type ProductUseCase struct {
	menuRepo  repository.MenuRepository
	auditRepo repository.AuditRepository
	queuePub  service.QueuePublisher
}

// NewProductUseCase creates a new ProductUseCase
func NewProductUseCase(
	menuRepo repository.MenuRepository,
	auditRepo repository.AuditRepository,
	queuePub service.QueuePublisher,
) *ProductUseCase {
	return &ProductUseCase{
		menuRepo:  menuRepo,
		auditRepo: auditRepo,
		queuePub:  queuePub,
	}
}

// UpdateProductStatus queues a product status update
func (uc *ProductUseCase) UpdateProductStatus(ctx context.Context, restaurantID, productID, newStatus, reason, userID string) error {
	// Get current status
	oldStatus, err := uc.menuRepo.GetProductStatus(ctx, restaurantID, productID)
	if err != nil {
		return fmt.Errorf("failed to get product status: %w", err)
	}

	// Publish event to queue
	event := &entity.ProductStatusChangeEvent{
		EventType:    entity.EventTypeProductStatusChanged,
		RestaurantID: restaurantID,
		ProductID:    productID,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Reason:       reason,
		Timestamp:    time.Now(),
		UserID:       userID,
	}

	if err := uc.queuePub.PublishProductStatusEvent(event); err != nil {
		return fmt.Errorf("failed to queue event: %w", err)
	}

	return nil
}

// ProcessProductStatusEvent processes a product status change event
func (uc *ProductUseCase) ProcessProductStatusEvent(ctx context.Context, event *entity.ProductStatusChangeEvent) error {
	// Update product status in DB
	oldStatus, err := uc.menuRepo.UpdateProductStatus(ctx, event.RestaurantID, event.ProductID, event.NewStatus)
	if err != nil {
		return fmt.Errorf("failed to update product status: %w", err)
	}

	// Use actual old status from DB
	if oldStatus != event.OldStatus {
		event.OldStatus = oldStatus
	}

	// Create audit record
	audit := &entity.ProductStatusAudit{
		ProductID: event.ProductID,
		EventType: event.EventType,
		OldStatus: event.OldStatus,
		NewStatus: event.NewStatus,
		Reason:    event.Reason,
		UserID:    event.UserID,
		Timestamp: event.Timestamp,
	}

	if err := uc.auditRepo.Create(ctx, audit); err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}

	return nil
}
