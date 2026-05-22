package http

import (
	"nop-go/services/vendorsvc/internal/server/http/handler"
	"nop-go/services/vendorsvc/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册供应商服务 HTTP 路由。
//
// 路由设计：
// - GET    /api/v1/vendors          — 供应商列表
// - POST   /api/v1/vendors          — 创建供应商
// - PUT    /api/v1/vendors/:id      — 更新供应商
// - DELETE /api/v1/vendors/:id      — 删除供应商
// - GET    /api/v1/vendors/apply    — 申请供应商（查询申请状态）
// - POST   /api/v1/vendors/apply    — 提交供应商申请
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	vendorHandler := handler.NewVendorHandler(services.Vendor)

	api := r.Group("/api/v1/vendors")
	{
		api.GET("", vendorHandler.List)              // 供应商列表
		api.POST("", vendorHandler.Create)           // 创建供应商
		api.PUT("/:id", vendorHandler.Update)        // 更新供应商
		api.DELETE("/:id", vendorHandler.Delete)     // 删除供应商
		api.GET("/apply", vendorHandler.GetApply)     // 申请供应商（查询申请状态）
		api.POST("/apply", vendorHandler.SubmitApply) // 提交供应商申请
	}
}