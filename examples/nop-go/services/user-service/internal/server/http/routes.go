package http

import (
	"nop-go/services/user-service/internal/server/http/handler"
	"nop-go/services/user-service/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册合并后用户服务的所有 HTTP 路由。
//
// 路由分组说明：
// - /api/v1/auth/*    认证相关端点（登录、注册）
// - /api/v1/users/*   用户信息相关端点（用户 CRUD、地址、外部关联、可下载产品）
// - /api/v1/gdprs/*   GDPR 请求相关端点（数据删除/导出请求管理）
//
// 中文说明：
// - Gin-first 模式：直接使用原生 Gin Engine 注册路由；
// - 抽象契约模式：使用 gorp.Router 抽象接口；
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	userHandler := handler.NewUserHandler(services.User)
	gdprHandler := handler.NewGdprHandler(services.Gdpr)

	// ---------- 认证相关路由 ----------
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", userHandler.Register) // 注册
		auth.POST("/login", userHandler.Login)       // 登录
	}

	// ---------- 用户信息相关路由 ----------
	// 注意：实际项目中 /api/v1/users/* 路由组需要认证中间件保护
	users := r.Group("/api/v1/users")
	{
		// 用户基本信息
		users.GET("", userHandler.ListUsers)       // 用户列表
		users.GET("/:id", userHandler.GetUser)     // 获取单个用户
		users.POST("", userHandler.CreateUser)     // 创建用户
		users.PUT("/:id", userHandler.UpdateUser)  // 更新用户信息
		users.DELETE("/:id", userHandler.DeleteUser) // 删除用户

		// 地址管理
		users.GET("/:id/addresses", userHandler.ListAddresses)       // 地址列表
		users.POST("/:id/addresses", userHandler.CreateAddress)      // 添加地址
		users.PUT("/addresses/:addrId", userHandler.UpdateAddress)   // 编辑地址
		users.DELETE("/addresses/:addrId", userHandler.DeleteAddress) // 删除地址

		// 外部关联
		users.GET("/:id/external-associations", userHandler.ListExternalAssociations)             // 外部关联列表
		users.POST("/:id/external-associations", userHandler.CreateExternalAssociation)           // 创建外部关联
		users.DELETE("/external-associations/:eaId", userHandler.DeleteExternalAssociation)       // 删除外部关联

		// 可下载产品
		users.GET("/:id/downloadable-products", userHandler.ListDownloadableProducts)             // 可下载产品列表
		users.POST("/:id/downloadable-products", userHandler.CreateDownloadableProduct)           // 创建可下载产品
		users.DELETE("/downloadable-products/:dpId", userHandler.DeleteDownloadableProduct)       // 删除可下载产品
	}

	// ---------- GDPR 相关路由 ----------
	gdprs := r.Group("/api/v1/gdprs")
	{
		gdprs.GET("", gdprHandler.List)       // GDPR 请求列表
		gdprs.GET("/:id", gdprHandler.GetByID) // 获取单个 GDPR 请求
		gdprs.POST("", gdprHandler.Create)     // 创建 GDPR 请求
		gdprs.PUT("/:id", gdprHandler.Update)  // 更新 GDPR 请求
		gdprs.DELETE("/:id", gdprHandler.Delete) // 删除 GDPR 请求
		gdprs.PUT("/:id/review", gdprHandler.Review)   // 审核 GDPR 请求
		gdprs.PUT("/:id/complete", gdprHandler.Complete) // 完成 GDPR 请求
	}
}