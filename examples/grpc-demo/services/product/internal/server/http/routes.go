package http

import (
	"grpc-demo/services/product/internal/server/http/handler"
	"grpc-demo/services/product/internal/service"

	gorp "github.com/ngq/gorp"
)

func RegisterRoutes(r gorp.Router, services *service.Services) {
	productHandler := handler.NewProductHandler(services.Product)

	api := r.Group("/api/v1/products")
	{
		api.GET("", productHandler.List)
		api.GET("/:id", productHandler.GetByID)
		api.POST("", productHandler.Create)
		api.DELETE("/:id", productHandler.Delete)
		api.POST("/:id/events/publish", productHandler.PublishStockChanged)
		api.POST("/events/consume-once", productHandler.ConsumeStockChangedOnce)
	}
}
