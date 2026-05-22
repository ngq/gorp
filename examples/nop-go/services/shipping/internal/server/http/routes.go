// Package http 提供 shipping 服务的 HTTP 路由注册
package http

import (
	"nop-go/services/shipping/internal/server/http/handler"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册配送服务的所有 HTTP 路由
// 路由分组: /api/v1/shipping
//
// 配送服务包含五个子领域：
// 1. 配送提供者（Provider）— 物流公司/承运商
// 2. 配送方式（Method）— 具体配送选项（如标准/加急）
// 3. 配送日期（DeliveryDate）— 可选配送日期
// 4. 仓库（Warehouse）— 发货仓库
// 5. 运费估算（Estimate）— 根据条件估算运费
func RegisterRoutes(r gorp.Router, h *handler.ShippingHandler) {
	// 配送提供者相关路由
	r.GET("/api/v1/shipping/providers", h.ListProviders)         // 配送提供者列表
	r.PUT("/api/v1/shipping/providers/:id", h.UpdateProvider)    // 更新配送提供者

	// 配送方式相关路由（CRUD）
	r.GET("/api/v1/shipping/methods", h.ListMethods)             // 配送方式列表
	r.POST("/api/v1/shipping/methods", h.CreateMethod)           // 创建配送方式
	r.PUT("/api/v1/shipping/methods/:id", h.UpdateMethod)        // 更新配送方式
	r.DELETE("/api/v1/shipping/methods/:id", h.DeleteMethod)     // 删除配送方式

	// 配送日期相关路由
	r.GET("/api/v1/shipping/delivery-dates", h.ListDeliveryDates)         // 配送日期列表
	r.POST("/api/v1/shipping/delivery-dates", h.CreateDeliveryDate)       // 创建配送日期
	r.PUT("/api/v1/shipping/delivery-dates/:id", h.UpdateDeliveryDate)    // 更新配送日期

	// 仓库相关路由
	r.GET("/api/v1/shipping/warehouses", h.ListWarehouses)        // 仓库列表
	r.POST("/api/v1/shipping/warehouses", h.CreateWarehouse)      // 创建仓库
	r.PUT("/api/v1/shipping/warehouses/:id", h.UpdateWarehouse)   // 更新仓库

	// 运费估算路由
	r.GET("/api/v1/shipping/estimate", h.EstimateShipping)        // 估算运费
}
