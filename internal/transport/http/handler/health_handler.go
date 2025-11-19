package handler

import (
	"net/http"

	"menu-parser/internal/usecase"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	healthUseCase *usecase.HealthUseCase
}

func NewHealthHandler(healthUseCase *usecase.HealthUseCase) *HealthHandler {
	return &HealthHandler{
		healthUseCase: healthUseCase,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	response := h.healthUseCase.Check(c.Request.Context())
	c.JSON(http.StatusOK, response)
}
