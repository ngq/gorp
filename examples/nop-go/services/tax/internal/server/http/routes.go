// Package http 提供 tax 服务的 HTTP 路由注册
package http

import (
	"nop-go/services/tax/internal/server/http/handler"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册税务服务的所有 HTTP 路由
// 路由分组: /api/v1/tax
//
// 税务服务包含两个子领域：
// 1. 税务提供者（Provider）— 税务计算引擎/服务
// 2. 税类别（Category）— 商品税分类
func RegisterRoutes(r gorp.Router, h *handler.TaxHandler) {
	// 税务提供者相关路由
	r.GET("/api/v1/tax/providers", h.ListProviders)         // 税务提供者列表
	r.PUT("/api/v1/tax/providers/:id", h.UpdateProvider)    // 更新税务提供者

	// 税类别相关路由（CRUD）
	r.GET("/api/v1/tax/categories", h.ListCategories)           // 税类别列表
	r.POST("/api/v1/tax/categories", h.CreateCategory)          // 创建税类别
	r.PUT("/api/v1/tax/categories/:id", h.UpdateCategory)       // 更新税类别
	r.DELETE("/api/v1/tax/categories/:id", h.DeleteCategory)    // 删除税类别
}