package http

import (
	"nop-go/services/gdpr/internal/server/http/handler"
	"nop-go/services/gdpr/internal/service"
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
	gdprHandler := handler.NewGdprHandler(services.Gdpr)

	api := r.Group("/api/v1/gdprs")
	{
		api.GET("", gdprHandler.List)
		api.GET("/:id", gdprHandler.GetByID)
		api.POST("", gdprHandler.Create)
		api.DELETE("/:id", gdprHandler.Delete)
	}
}
