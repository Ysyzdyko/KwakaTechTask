package repository

import (
	"context"
	"fmt"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/repository"
	"menu-parser/pkg/database"
)

type AuditRepository struct {
	db *database.MongoDB
}

func NewAuditRepository(db *database.MongoDB) repository.AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, audit *entity.ProductStatusAudit) error {
	_, err := r.db.Database.Collection("product_status_audit").InsertOne(ctx, audit)
	if err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}
	return nil
}


