package http

import (
	"nop-go/services/logging/internal/server/http/handler"
	"nop-go/services/logging/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 HTTP 路由。
//
// 中文说明：
// - Gin-first 模式：直接使用原生 Gin Engine 注册路由；
// - 抽象契约模式：使用 gorp.Router 抽象接口；
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	loggingHandler := handler.NewLoggingHandler(services.Logging)

	api := r.Group("/api/v1/loggings")
	{
		api.GET("", loggingHandler.List)
		api.GET("/:id", loggingHandler.GetByID)
		api.POST("", loggingHandler.Create)
		api.DELETE("/:id", loggingHandler.Delete)
	}
}
