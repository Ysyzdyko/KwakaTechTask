package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ParsingTaskStatus string

const (
	TaskStatusQueued     ParsingTaskStatus = "queued"
	TaskStatusProcessing ParsingTaskStatus = "processing"
	TaskStatusCompleted  ParsingTaskStatus = "completed"
	TaskStatusFailed     ParsingTaskStatus = "failed"
)

type ParsingTask struct {
	ID             string              `json:"task_id" bson:"_id"`
	Status         ParsingTaskStatus   `json:"status" bson:"status"`
	SpreadsheetID  string              `json:"spreadsheet_id" bson:"spreadsheet_id"`
	RestaurantName string              `json:"restaurant_name" bson:"restaurant_name"`
	MenuID         *primitive.ObjectID `json:"menu_id,omitempty" bson:"menu_id,omitempty"`
	ErrorMessage   string              `json:"error,omitempty" bson:"error_message,omitempty"`
	RetryCount     int                 `json:"retry_count" bson:"retry_count"`
	CreatedAt      time.Time           `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at" bson:"updated_at"`
}


