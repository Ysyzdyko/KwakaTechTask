package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"menu-parser/internal/domain/entity"
	"menu-parser/internal/domain/repository"
	"menu-parser/internal/domain/service"
)

// MenuUseCase handles menu-related business logic
type MenuUseCase struct {
	menuRepo     repository.MenuRepository
	taskRepo     repository.TaskRepository
	parser       service.SheetsParser
	queuePub     service.QueuePublisher
}

// NewMenuUseCase creates a new MenuUseCase
func NewMenuUseCase(
	menuRepo repository.MenuRepository,
	taskRepo repository.TaskRepository,
	parser service.SheetsParser,
	queuePub service.QueuePublisher,
) *MenuUseCase {
	return &MenuUseCase{
		menuRepo: menuRepo,
		taskRepo: taskRepo,
		parser:   parser,
		queuePub: queuePub,
	}
}

// CreateParsingTask creates a new parsing task and queues it
func (uc *MenuUseCase) CreateParsingTask(ctx context.Context, spreadsheetID, restaurantName string) (string, error) {
	taskID := uuid.New().String()
	
	task := &entity.ParsingTask{
		ID:             taskID,
		Status:         entity.TaskStatusQueued,
		SpreadsheetID:  spreadsheetID,
		RestaurantName: restaurantName,
		RetryCount:     0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := uc.taskRepo.Create(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	if err := uc.queuePub.PublishMenuParsingTask(taskID); err != nil {
		return "", fmt.Errorf("failed to queue task: %w", err)
	}

	return taskID, nil
}

// GetTaskStatus retrieves the status of a parsing task
func (uc *MenuUseCase) GetTaskStatus(ctx context.Context, taskID string) (*entity.ParsingTask, error) {
	return uc.taskRepo.GetByID(ctx, taskID)
}

// GetMenu retrieves a menu by ID
func (uc *MenuUseCase) GetMenu(ctx context.Context, menuID string) (*entity.Menu, error) {
	return uc.menuRepo.GetByID(ctx, menuID)
}

// ProcessMenuParsing processes a menu parsing task
func (uc *MenuUseCase) ProcessMenuParsing(ctx context.Context, taskID string) error {
	task, err := uc.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Update status to processing
	if err := uc.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusProcessing, nil, ""); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	// Parse menu
	menu, err := uc.parser.ParseMenu(ctx, task.SpreadsheetID, task.RestaurantName)
	if err != nil {
		uc.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusFailed, nil, err.Error())
		return fmt.Errorf("failed to parse menu: %w", err)
	}

	// Save menu
	savedMenu, err := uc.menuRepo.Create(ctx, menu)
	if err != nil {
		uc.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusFailed, nil, err.Error())
		return fmt.Errorf("failed to save menu: %w", err)
	}

	// Update task status to completed
	if err := uc.taskRepo.UpdateStatus(ctx, taskID, entity.TaskStatusCompleted, &savedMenu.ID, ""); err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

