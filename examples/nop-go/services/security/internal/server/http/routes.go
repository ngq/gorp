package http

import (
	"nop-go/services/security/internal/server/http/handler"
	"nop-go/services/security/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册安全服务 HTTP 路由。
//
// 路由设计：
// 权限管理：
// - GET    /api/v1/permissions         — 权限列表
// - POST   /api/v1/permissions         — 创建权限
// - PUT    /api/v1/permissions/:id     — 更新权限
// - DELETE /api/v1/permissions/:id     — 删除权限
// ACL管理：
// - GET    /api/v1/acl                — ACL记录列表
// - POST   /api/v1/acl                — 创建ACL记录
// - DELETE /api/v1/acl/:id            — 删除ACL记录
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	permHandler := handler.NewPermissionHandler(services.Permission)
	aclHandler := handler.NewACLHandler(services.ACL)

	// 权限相关路由
	perm := r.Group("/api/v1/permissions")
	{
		perm.GET("", permHandler.List)              // 权限列表
		perm.POST("", permHandler.Create)           // 创建权限
		perm.PUT("/:id", permHandler.Update)        // 更新权限
		perm.DELETE("/:id", permHandler.Delete)     // 删除权限
	}

	// ACL相关路由
	acl := r.Group("/api/v1/acl")
	{
		acl.GET("", aclHandler.List)              // ACL记录列表
		acl.POST("", aclHandler.Create)           // 创建ACL记录
		acl.DELETE("/:id", aclHandler.Delete)     // 删除ACL记录
	}
}