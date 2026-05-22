// Package http 提供 discount 服务的 HTTP 路由注册
package http

import (
	"nop-go/services/discount/internal/server/http/handler"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册折扣服务的所有 HTTP 路由
// 路由分组: /api/v1/discounts
//
// 折扣服务包含以下子领域：
// 1. 折扣（Discount）— 折扣主体 CRUD
// 2. 折扣关联商品（Products）— 折扣适用的商品
// 3. 折扣关联分类（Categories）— 折扣适用的分类
// 4. 折扣关联制造商（Manufacturers）— 折扣适用的制造商
// 5. 折扣使用历史（UsageHistory）— 折扣使用记录
func RegisterRoutes(r gorp.Router, h *handler.DiscountHandler) {
	// 折扣 CRUD 路由
	r.GET("/api/v1/discounts", h.ListDiscounts)               // 折扣列表
	r.POST("/api/v1/discounts", h.CreateDiscount)              // 创建折扣
	r.PUT("/api/v1/discounts/:id", h.UpdateDiscount)           // 更新折扣
	r.DELETE("/api/v1/discounts/:id", h.DeleteDiscount)        // 删除折扣

	// 折扣关联子资源路由
	r.GET("/api/v1/discounts/:id/products", h.ListDiscountProducts)       // 折扣关联商品
	r.GET("/api/v1/discounts/:id/categories", h.ListDiscountCategories)   // 折扣关联分类
	r.GET("/api/v1/discounts/:id/manufacturers", h.ListDiscountManufacturers) // 折扣关联制造商
	r.GET("/api/v1/discounts/:id/usage-history", h.ListDiscountUsageHistory) // 折扣使用历史
}