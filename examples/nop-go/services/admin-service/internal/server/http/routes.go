// Package http 路由注册 —— admin-service 的统一 HTTP 路由入口
//
// 合并6个子模块的路由到统一的注册函数：
// 1. /api/v1/permissions — 权限管理（来自 security）
// 2. /api/v1/acl — 访问控制列表（来自 security）
// 3. /api/v1/plugins — 插件管理（来自 plugin）
// 4. /api/v1/stores — 门店管理（来自 store）
// 5. /api/v1/activity-logs — 活动日志（来自 logging）
// 6. /api/v1/system-logs — 系统日志（来自 logging）
// 7. /api/v1/discounts — 优惠管理（来自 discount）
// 8. /api/v1/discounts/:id/usages — 优惠使用记录（来自 discount）
// 9. /api/v1/vendors — 供应商管理（来自 vendorsvc）
package http

import (
	"nop-go/services/admin-service/internal/server/http/handler"
	"nop-go/services/admin-service/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册管理后台统一 HTTP 路由
//
// 路由分组设计：
// - 权限管理：GET/POST /permissions, PUT/DELETE /permissions/:id
// - ACL管理：GET/POST /acl, PUT/DELETE /acl/:id
// - 插件管理：GET/POST /plugins, PUT/DELETE /plugins/:id
// - 门店管理：GET/POST /stores, PUT/DELETE /stores/:id
// - 活动日志：GET /activity-logs, GET /activity-logs/:id
// - 系统日志：GET /system-logs, GET /system-logs/:id
// - 优惠管理：GET/POST /discounts, PUT/DELETE /discounts/:id
// - 优惠使用记录：GET /discounts/:id/usages
// - 供应商管理：GET/POST /vendors, PUT/DELETE /vendors/:id
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	// 权限管理路由
	permHandler := handler.NewPermissionHandler(services.Security)
	perm := r.Group("/api/v1/permissions")
	{
		perm.GET("", permHandler.List)              // 权限列表
		perm.POST("", permHandler.Create)           // 创建权限
		perm.GET("/:id", permHandler.Get)           // 获取权限详情
		perm.PUT("/:id", permHandler.Update)        // 更新权限
		perm.DELETE("/:id", permHandler.Delete)     // 删除权限
	}

	// ACL管理路由
	aclHandler := handler.NewACLHandler(services.Security)
	acl := r.Group("/api/v1/acl")
	{
		acl.GET("", aclHandler.List)               // ACL列表
		acl.POST("", aclHandler.Create)            // 创建ACL
		acl.GET("/:id", aclHandler.Get)            // 获取ACL详情
		acl.PUT("/:id", aclHandler.Update)         // 更新ACL
		acl.DELETE("/:id", aclHandler.Delete)      // 删除ACL
	}

	// 插件管理路由
	pluginHandler := handler.NewPluginHandler(services.Plugin)
	plugin := r.Group("/api/v1/plugins")
	{
		plugin.GET("", pluginHandler.List)           // 插件列表
		plugin.POST("", pluginHandler.Create)        // 创建插件
		plugin.GET("/:id", pluginHandler.Get)        // 获取插件详情
		plugin.PUT("/:id", pluginHandler.Update)     // 更新插件
		plugin.DELETE("/:id", pluginHandler.Delete)  // 删除插件
	}

	// 门店管理路由
	storeHandler := handler.NewStoreHandler(services.Store)
	store := r.Group("/api/v1/stores")
	{
		store.GET("", storeHandler.List)            // 门店列表
		store.POST("", storeHandler.Create)         // 创建门店
		store.GET("/:id", storeHandler.Get)         // 获取门店详情
		store.PUT("/:id", storeHandler.Update)      // 更新门店
		store.DELETE("/:id", storeHandler.Delete)   // 删除门店
	}

	// 活动日志路由
	activityHandler := handler.NewActivityLogHandler(services.Logging)
	activity := r.Group("/api/v1/activity-logs")
	{
		activity.GET("", activityHandler.List)      // 活动日志列表
		activity.GET("/:id", activityHandler.Get)   // 获取活动日志详情
		activity.POST("", activityHandler.Create)   // 创建活动日志
	}

	// 系统日志路由
	systemHandler := handler.NewSystemLogHandler(services.Logging)
	system := r.Group("/api/v1/system-logs")
	{
		system.GET("", systemHandler.List)          // 系统日志列表
		system.GET("/:id", systemHandler.Get)       // 获取系统日志详情
		system.POST("", systemHandler.Create)       // 创建系统日志
	}

	// 优惠管理路由
	discountHandler := handler.NewDiscountHandler(services.Discount)
	discount := r.Group("/api/v1/discounts")
	{
		discount.GET("", discountHandler.List)               // 优惠列表
		discount.POST("", discountHandler.Create)            // 创建优惠
		discount.GET("/:id", discountHandler.Get)            // 获取优惠详情
		discount.PUT("/:id", discountHandler.Update)         // 更新优惠
		discount.DELETE("/:id", discountHandler.Delete)      // 删除优惠
		// 优惠使用记录子资源
		discount.GET("/:id/usages", discountHandler.ListUsages) // 优惠使用记录列表
	}

	// 供应商管理路由
	vendorHandler := handler.NewVendorHandler(services.Vendor)
	vendor := r.Group("/api/v1/vendors")
	{
		vendor.GET("", vendorHandler.List)             // 供应商列表
		vendor.POST("", vendorHandler.Create)          // 创建供应商
		vendor.GET("/:id", vendorHandler.Get)          // 获取供应商详情
		vendor.PUT("/:id", vendorHandler.Update)       // 更新供应商
		vendor.DELETE("/:id", vendorHandler.Delete)    // 删除供应商
	}
}