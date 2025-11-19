package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Menu struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name             string             `json:"name" bson:"name"`
	RestaurantID     string             `json:"restaurant_id" bson:"restaurant_id"`
	Products         []Product          `json:"products" bson:"products"`
	AttributesGroups []AttributesGroup  `json:"attributes_groups" bson:"attributes_groups"`
	Attributes       []Attribute        `json:"attributes" bson:"attributes"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at" bson:"updated_at"`
}

type Product struct {
	ExtID      string                 `json:"ext_id" bson:"ext_id"`
	Name       string                 `json:"name" bson:"name"`
	Price      float64                `json:"price" bson:"price"`
	PriceOld   float64                `json:"price_old,omitempty" bson:"price_old,omitempty"`
	Status     string                 `json:"status" bson:"status"`
	Attributes map[string]interface{} `json:"attributes,omitempty" bson:"attributes,omitempty"`
}

type AttributesGroup struct {
	ID         string      `json:"id" bson:"id"`
	Name       string      `json:"name" bson:"name"`
	Attributes []Attribute `json:"attributes" bson:"attributes"`
	IsRequired bool        `json:"is_required" bson:"is_required"`
}

type Attribute struct {
	ID    string `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	Value string `json:"value,omitempty" bson:"value,omitempty"`
}


