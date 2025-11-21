package http

import (
	"menu-parser/internal/transport/http/handler"
	"menu-parser/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupRouter(
	menuUseCase *usecase.MenuUseCase,
	productUseCase *usecase.ProductUseCase,
	healthUseCase *usecase.HealthUseCase,
) *gin.Engine {
	router := gin.Default()

	menuHandler := handler.NewMenuHandler(menuUseCase)
	productHandler := handler.NewProductHandler(productUseCase)
	healthHandler := handler.NewHealthHandler(healthUseCase)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/parse", menuHandler.ParseMenu)
		v1.GET("/parse/:task_id", menuHandler.GetTaskStatus)
		v1.GET("/menu/:menu_id", menuHandler.GetMenu)
		v1.PATCH("/restaurants/:restaurant_id/products/:product_id/status", productHandler.UpdateProductStatus)
		v1.GET("/health", healthHandler.HealthCheck)
	}

	return router
}
