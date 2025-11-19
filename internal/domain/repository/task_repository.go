package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"menu-parser/internal/domain/entity"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.ParsingTask) error
	GetByID(ctx context.Context, taskID string) (*entity.ParsingTask, error)
	UpdateStatus(ctx context.Context, taskID string, status entity.ParsingTaskStatus, menuID *primitive.ObjectID, errorMsg string) error
	IncrementRetryCount(ctx context.Context, taskID string) error
}


