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

type MenuRepository struct {
	db *database.MongoDB
}

func NewMenuRepository(db *database.MongoDB) repository.MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) Create(ctx context.Context, menu *entity.Menu) (*entity.Menu, error) {
	menu.CreatedAt = time.Now()
	menu.UpdatedAt = time.Now()

	result, err := r.db.Database.Collection("menus").InsertOne(ctx, menu)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu: %w", err)
	}

	menu.ID = result.InsertedID.(primitive.ObjectID)
	return menu, nil
}

func (r *MenuRepository) GetByID(ctx context.Context, menuID string) (*entity.Menu, error) {
	objectID, err := primitive.ObjectIDFromHex(menuID)
	if err != nil {
		return nil, fmt.Errorf("invalid menu ID: %w", err)
	}

	var menu entity.Menu
	err = r.db.Database.Collection("menus").FindOne(ctx, bson.M{"_id": objectID}).Decode(&menu)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("menu not found")
		}
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	return &menu, nil
}

func (r *MenuRepository) GetProductStatus(ctx context.Context, productID string) (string, error) {
	var menu entity.Menu
	filter := bson.M{"products.ext_id": productID}

	err := r.db.Database.Collection("menus").FindOne(ctx, filter).Decode(&menu)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("product not found")
		}
		return "", fmt.Errorf("failed to find product: %w", err)
	}

	for _, product := range menu.Products {
		if product.ExtID == productID {
			return product.Status, nil
		}
	}

	return "", fmt.Errorf("product not found in menu")
}

func (r *MenuRepository) UpdateProductStatus(ctx context.Context, productID, newStatus string) (string, error) {
	var menu entity.Menu
	filter := bson.M{"products.ext_id": productID}

	err := r.db.Database.Collection("menus").FindOne(ctx, filter).Decode(&menu)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", fmt.Errorf("product not found")
		}
		return "", fmt.Errorf("failed to find product: %w", err)
	}

	var oldStatus string
	for i, product := range menu.Products {
		if product.ExtID == productID {
			oldStatus = product.Status
			menu.Products[i].Status = newStatus
			menu.UpdatedAt = time.Now()
			break
		}
	}

	if oldStatus == "" {
		return "", fmt.Errorf("product not found in menu")
	}

	_, err = r.db.Database.Collection("menus").UpdateOne(
		ctx,
		bson.M{"_id": menu.ID},
		bson.M{"$set": bson.M{
			"products":   menu.Products,
			"updated_at": menu.UpdatedAt,
		}},
	)
	if err != nil {
		return "", fmt.Errorf("failed to update product status: %w", err)
	}

	return oldStatus, nil
}


