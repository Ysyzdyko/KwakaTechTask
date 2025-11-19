package repository

import (
	"context"

	"menu-parser/internal/domain/entity"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *entity.Menu) (*entity.Menu, error)
	GetByID(ctx context.Context, menuID string) (*entity.Menu, error)
	GetProductStatus(ctx context.Context, productID string) (string, error)
	UpdateProductStatus(ctx context.Context, productID, newStatus string) (string, error)
}


