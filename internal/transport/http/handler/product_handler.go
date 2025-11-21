package handler

import (
	"net/http"

	"menu-parser/internal/transport/http/dto"
	"menu-parser/internal/usecase"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productUseCase *usecase.ProductUseCase
}

func NewProductHandler(productUseCase *usecase.ProductUseCase) *ProductHandler {
	return &ProductHandler{
		productUseCase: productUseCase,
	}
}

func (h *ProductHandler) UpdateProductStatus(c *gin.Context) {
	restaurantID := c.Param("restaurant_id")
	productID := c.Param("product_id")

	var req dto.ProductStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	if userID == "" {
		userID = "system"
	}

	err := h.productUseCase.UpdateProductStatus(c.Request.Context(), restaurantID, productID, req.Status, req.Reason, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, dto.ProductStatusUpdateResponse{
		Success: true,
		Message: "Status update queued",
	})
}
