package handler

import (
	"net/http"

	"menu-parser/internal/transport/http/dto"
	"menu-parser/internal/usecase"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	menuUseCase *usecase.MenuUseCase
}

func NewMenuHandler(menuUseCase *usecase.MenuUseCase) *MenuHandler {
	return &MenuHandler{
		menuUseCase: menuUseCase,
	}
}

func (h *MenuHandler) ParseMenu(c *gin.Context) {
	var req dto.ParseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	taskID, err := h.menuUseCase.CreateParsingTask(c.Request.Context(), req.SpreadsheetID, req.RestaurantName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ParseResponse{
		TaskID: taskID,
		Status: "queued",
	})
}

func (h *MenuHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.menuUseCase.GetTaskStatus(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToTaskStatusResponse(task))
}

func (h *MenuHandler) GetMenu(c *gin.Context) {
	menuID := c.Param("menu_id")

	menu, err := h.menuUseCase.GetMenu(c.Request.Context(), menuID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToMenuResponse(menu))
}
