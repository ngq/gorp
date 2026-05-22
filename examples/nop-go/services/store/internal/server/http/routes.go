package http

import (
	"nop-go/services/store/internal/server/http/handler"
	"nop-go/services/store/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册店铺服务 HTTP 路由。
//
// 路由设计：
// - GET    /api/v1/stores      — 店铺列表
// - POST   /api/v1/stores      — 创建店铺
// - PUT    /api/v1/stores/:id  — 更新店铺
// - DELETE /api/v1/stores/:id  — 删除店铺
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	storeHandler := handler.NewStoreHandler(services.Store)

	api := r.Group("/api/v1/stores")
	{
		api.GET("", storeHandler.List)              // 店铺列表
		api.POST("", storeHandler.Create)           // 创建店铺
		api.PUT("/:id", storeHandler.Update)        // 更新店铺
		api.DELETE("/:id", storeHandler.Delete)     // 删除店铺
	}
}