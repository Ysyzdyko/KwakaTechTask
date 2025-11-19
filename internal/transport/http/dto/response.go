package dto

import (
	"time"

	"menu-parser/internal/domain/entity"
)

type ParseResponse struct {
	TaskID string `json:"task_id"`
	Status string `json:"status"`
}

type TaskStatusResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	MenuID    string    `json:"menu_id,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProductStatusUpdateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type MenuResponse struct {
	ID               string                   `json:"_id"`
	Name             string                   `json:"name"`
	RestaurantID     string                   `json:"restaurant_id"`
	Products         []entity.Product         `json:"products"`
	AttributesGroups []entity.AttributesGroup `json:"attributes_groups"`
	Attributes       []entity.Attribute       `json:"attributes"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

func ToMenuResponse(menu *entity.Menu) *MenuResponse {
	return &MenuResponse{
		ID:               menu.ID.Hex(),
		Name:             menu.Name,
		RestaurantID:     menu.RestaurantID,
		Products:         menu.Products,
		AttributesGroups: menu.AttributesGroups,
		Attributes:       menu.Attributes,
		CreatedAt:        menu.CreatedAt,
		UpdatedAt:        menu.UpdatedAt,
	}
}

func ToTaskStatusResponse(task *entity.ParsingTask) *TaskStatusResponse {
	resp := &TaskStatusResponse{
		TaskID:    task.ID,
		Status:    string(task.Status),
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
	}

	if task.MenuID != nil {
		resp.MenuID = task.MenuID.Hex()
	}

	if task.ErrorMessage != "" {
		resp.Error = task.ErrorMessage
	}

	return resp
}
