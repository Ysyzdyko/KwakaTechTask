package database

import (
	"context"
	"fmt"
	"time"

	"menu-parser/pkg/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().
		ApplyURI(cfg.MongoDBURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	db := client.Database(cfg.MongoDBDatabase)

	// Create indexes
	if err := createIndexes(ctx, db); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return &MongoDB{
		Client:   client,
		Database: db,
	}, nil
}

func createIndexes(ctx context.Context, db *mongo.Database) error {
	// Indexes for menus collection
	menusCollection := db.Collection("menus")
	menusIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"restaurant_id": 1},
		},
		{
			Keys: map[string]interface{}{"products.ext_id": 1},
		},
	}
	if _, err := menusCollection.Indexes().CreateMany(ctx, menusIndexes); err != nil {
		return fmt.Errorf("failed to create menus indexes: %w", err)
	}

	// Indexes for parsing_tasks collection
	tasksCollection := db.Collection("parsing_tasks")
	tasksIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"_id": 1},
		},
		{
			Keys: map[string]interface{}{"status": 1},
		},
		{
			Keys: map[string]interface{}{"created_at": 1},
		},
	}
	if _, err := tasksCollection.Indexes().CreateMany(ctx, tasksIndexes); err != nil {
		return fmt.Errorf("failed to create parsing_tasks indexes: %w", err)
	}

	// Indexes for product_status_audit collection
	auditCollection := db.Collection("product_status_audit")
	auditIndexes := []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"product_id": 1},
		},
		{
			Keys: map[string]interface{}{"timestamp": 1},
		},
	}
	if _, err := auditCollection.Indexes().CreateMany(ctx, auditIndexes); err != nil {
		return fmt.Errorf("failed to create product_status_audit indexes: %w", err)
	}

	return nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.Client.Ping(ctx, nil)
}


