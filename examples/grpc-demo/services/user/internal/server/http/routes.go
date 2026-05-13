package http

import (
	"grpc-demo/services/user/internal/server/http/handler"
	"grpc-demo/services/user/internal/service"

	gorp "github.com/ngq/gorp"
)

func RegisterRoutes(r gorp.HTTPRouter, services *service.Services) {
	userHandler := handler.NewUserHandler(services.User)

	api := r.Group("/api/v1/users")
	{
		api.GET("", userHandler.List)
		api.GET("/:id", userHandler.GetByID)
		api.POST("", userHandler.Create)
		api.DELETE("/:id", userHandler.Delete)
	}
}
