package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductEventType string

const (
	EventTypeProductCreated       ProductEventType = "product.created"
	EventTypeProductUpdated       ProductEventType = "product.updated"
	EventTypeProductStatusChanged ProductEventType = "product.status_changed"
	EventTypeProductDeleted       ProductEventType = "product.deleted"
)

type ProductStatus string

const (
	ProductStatusAvailable    ProductStatus = "available"
	ProductStatusNotAvailable ProductStatus = "not_available"
	ProductStatusDeleted      ProductStatus = "deleted"
)

type ProductStatusAudit struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	ProductID string             `json:"product_id" bson:"product_id"`
	EventType ProductEventType   `json:"event_type" bson:"event_type"`
	OldStatus string             `json:"old_status" bson:"old_status"`
	NewStatus string             `json:"new_status" bson:"new_status"`
	Reason    string             `json:"reason" bson:"reason"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

type ProductStatusChangeEvent struct {
	EventType    ProductEventType `json:"event_type"`
	RestaurantID string           `json:"restaurant_id"`
	ProductID    string           `json:"product_id"`
	OldStatus    string           `json:"old_status"`
	NewStatus    string           `json:"new_status"`
	Reason       string           `json:"reason"`
	Timestamp    time.Time        `json:"timestamp"`
	UserID       string           `json:"user_id"`
}
