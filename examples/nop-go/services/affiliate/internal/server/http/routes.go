package http

import (
	"nop-go/services/affiliate/internal/server/http/handler"
	"nop-go/services/affiliate/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册联盟服务 HTTP 路由。
//
// 路由设计：
// - GET    /api/v1/affiliates             — 联盟列表
// - POST   /api/v1/affiliates             — 创建联盟
// - PUT    /api/v1/affiliates/:id         — 更新联盟
// - DELETE /api/v1/affiliates/:id         — 删除联盟
// - GET    /api/v1/affiliates/:id/orders  — 联盟关联订单
// - GET    /api/v1/affiliates/:id/customers — 联盟关联客户
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	affHandler := handler.NewAffiliateHandler(services.Affiliate)

	api := r.Group("/api/v1/affiliates")
	{
		api.GET("", affHandler.List)                   // 联盟列表
		api.POST("", affHandler.Create)                // 创建联盟
		api.PUT("/:id", affHandler.Update)             // 更新联盟
		api.DELETE("/:id", affHandler.Delete)          // 删除联盟
		api.GET("/:id/orders", affHandler.ListOrders)     // 联盟关联订单
		api.GET("/:id/customers", affHandler.ListCustomers) // 联盟关联客户
	}
}