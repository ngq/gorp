package http

import (
	"admin/app/http/handler"
	"admin/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册 HTTP 路由。
//
// 中文说明：
// - 使用框架提供的 HTTPRouter 抽象接口；
// - 调用 internal/service 层注册业务路由；
// - `/healthz` 与 `/metrics` 等基础端点由 application 默认主线统一注册，这里只关注业务 API；
// - 业务是否启用认证或缓存能力，仍由项目自行决定；
// - 使用 gorp.HTTPContext 接口，与模板生成代码风格一致。
func RegisterRoutes(r gorp.HTTPRouter, svc *service.Services) {
	// Demo 处理器（使用模板生成的 HTTPContext 风格）
	demoHandler := handler.NewDemoHandler(svc.Demo)
	authHandler := handler.NewAuthHandler(svc.Auth)
	userHandler := handler.NewUserHandler(svc.User)
	roleHandler := handler.NewRoleHandler(svc.Role)

	api := r.Group("/api/v1")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
		}

		// Demo 路由（无需认证）
		demos := api.Group("/demos")
		{
			demos.POST("", demoHandler.Create)
			demos.GET("/:id", demoHandler.GetByID)
			demos.GET("", demoHandler.List)
			demos.PUT("/:id", demoHandler.Update)
			demos.DELETE("/:id", demoHandler.Delete)
		}

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(handler.AuthMiddleware(svc.Auth))
		{
			// 当前用户
			protected.GET("/auth/me", authHandler.GetCurrentUser)

			// 用户管理
			users := protected.Group("/users")
			{
				users.GET("", userHandler.List)
				users.POST("", userHandler.Create)
				users.PUT("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
				users.POST("/:id/roles", userHandler.AssignRoles)
			}

			// 角色管理
			roles := protected.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.POST("", roleHandler.Create)
				roles.DELETE("/:id", roleHandler.Delete)
			}
		}
	}
}
