package http

import (
	"grpc-demo/services/order/internal/server/http/handler"
	"grpc-demo/services/order/internal/service"

	gorp "github.com/ngq/gorp"
)

func RegisterRoutes(r gorp.HTTPRouter, services *service.Services) {
	orderHandler := handler.NewOrderHandler(services.Order)

	api := r.Group("/api/v1/orders")
	{
		api.GET("", orderHandler.List)
		api.GET("/:id", orderHandler.GetByID)
		api.POST("", orderHandler.Create)
		api.DELETE("/:id", orderHandler.Delete)
		api.GET("/:id/user", orderHandler.GetOrderUser)
		api.POST("/:id/lock-demo", orderHandler.LockOrderDemo)
	}
}
