package repository

import (
	"context"
	"fmt"
	"time"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/repository"
	"menu-parser/pkg/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskRepository struct {
	db *database.MongoDB
}

func NewTaskRepository(db *database.MongoDB) repository.TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *entity.ParsingTask) error {
	_, err := r.db.Database.Collection("parsing_tasks").InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create parsing task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, taskID string) (*entity.ParsingTask, error) {
	var task entity.ParsingTask
	err := r.db.Database.Collection("parsing_tasks").FindOne(ctx, bson.M{"_id": taskID}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get parsing task: %w", err)
	}
	return &task, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID string, status entity.ParsingTaskStatus, menuID *primitive.ObjectID, errorMsg string) error {
	update := bson.M{
		"$set": bson.M{
			"status":     string(status),
			"updated_at": time.Now(),
		},
	}

	if menuID != nil {
		update["$set"].(bson.M)["menu_id"] = menuID
	}

	if errorMsg != "" {
		update["$set"].(bson.M)["error_message"] = errorMsg
	}

	_, err := r.db.Database.Collection("parsing_tasks").UpdateOne(
		ctx,
		bson.M{"_id": taskID},
		update,
	)
	if err != nil {
		return fmt.Errorf("failed to update parsing task: %w", err)
	}
	return nil
}

func (r *TaskRepository) IncrementRetryCount(ctx context.Context, taskID string) error {
	_, err := r.db.Database.Collection("parsing_tasks").UpdateOne(
		ctx,
		bson.M{"_id": taskID},
		bson.M{
			"$inc": bson.M{"retry_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}


