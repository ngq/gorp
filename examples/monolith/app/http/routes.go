package http

import (
	"monolith/app/http/handler"
	"monolith/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 HTTP 路由。
//
// 中文说明：
// - 使用框架提供的 Router；
// - 调用 internal/service 层注册业务路由；
// - `/healthz` 与 `/metrics` 等基础端点由 application 默认主线统一注册，这里只关注业务 API；
// - 业务是否启用认证或缓存能力，仍由项目自行决定；
func RegisterRoutes(r gorp.Router, svc *service.Services) {
	demoHandler := handler.NewDemoHandler(svc.Demo)

	api := r.Group("/api/v1")
	{
		demos := api.Group("/demos")
		{
			demos.POST("", demoHandler.Create)
			demos.GET("/:id", demoHandler.GetByID)
			demos.GET("", demoHandler.List)
			demos.PUT("/:id", demoHandler.Update)
			demos.DELETE("/:id", demoHandler.Delete)
		}
	}
}
