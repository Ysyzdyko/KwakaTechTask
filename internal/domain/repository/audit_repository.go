package repository

import (
	"context"

	"menu-parser/internal/domain/entity"
)

type AuditRepository interface {
	Create(ctx context.Context, audit *entity.ProductStatusAudit) error
}


