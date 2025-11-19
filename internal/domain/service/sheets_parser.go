package service

import (
	"context"

	"menu-parser/internal/domain/entity"
)

type SheetsParser interface {
	ParseMenu(ctx context.Context, spreadsheetID, restaurantName string) (*entity.Menu, error)
}


