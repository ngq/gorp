package http

import (
	"nop-go/services/plugin/internal/server/http/handler"
	"nop-go/services/plugin/internal/service"
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
	pluginHandler := handler.NewPluginHandler(services.Plugin)

	api := r.Group("/api/v1/plugins")
	{
		api.GET("", pluginHandler.List)
		api.GET("/:id", pluginHandler.GetByID)
		api.POST("", pluginHandler.Create)
		api.DELETE("/:id", pluginHandler.Delete)
	}
}
