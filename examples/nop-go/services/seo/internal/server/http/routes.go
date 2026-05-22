package http

import (
	"nop-go/services/seo/internal/server/http/handler"
	"nop-go/services/seo/internal/service"
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
	seoHandler := handler.NewSeoHandler(services.Seo)

	api := r.Group("/api/v1/seos")
	{
		api.GET("", seoHandler.List)
		api.GET("/:id", seoHandler.GetByID)
		api.POST("", seoHandler.Create)
		api.DELETE("/:id", seoHandler.Delete)
	}
}
