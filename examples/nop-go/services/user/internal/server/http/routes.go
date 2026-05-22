package http

import (
	"nop-go/services/user/internal/server/http/handler"
	"nop-go/services/user/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册用户服务的所有 HTTP 路由。
//
// 路由分组说明：
// - /api/v1/auth/*    认证相关端点（登录、登出、注册、密码恢复、MFA）
// - /api/v1/users/*   用户信息相关端点（个人信息、地址、头像、可下载产品等）
//
// 中文说明：
// - Gin-first 模式：直接使用原生 Gin Engine 注册路由；
// - 抽象契约模式：使用 gorp.Router 抽象接口；
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	userHandler := handler.NewUserHandler(services.User)

	// ---------- 认证相关路由 ----------
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/login", userHandler.Login)                       // 登录
		auth.POST("/logout", userHandler.Logout)                     // 登出
		auth.POST("/register", userHandler.Register)                 // 注册
		auth.POST("/password-recovery", userHandler.PasswordRecovery)                         // 密码恢复
		auth.POST("/password-recovery/confirm", userHandler.ConfirmPasswordRecovery)           // 确认密码恢复
		auth.GET("/multi-factor-verification", userHandler.MultiFactorVerification)            // 多因素验证
	}

	// ---------- 用户信息相关路由 ----------
	// 注意：实际项目中 /api/v1/users/* 路由组需要认证中间件保护
	users := r.Group("/api/v1/users")
	{
		// 用户基本信息
		users.GET("/info", userHandler.GetUserInfo)       // 获取客户信息
		users.PUT("/info", userHandler.UpdateUserInfo)    // 更新客户信息
		users.PUT("/password", userHandler.ChangePassword) // 修改密码

		// 地址管理
		users.GET("/addresses", userHandler.ListAddresses)                  // 地址列表
		users.POST("/addresses", userHandler.AddAddress)                   // 添加地址
		users.PUT("/addresses/:id", userHandler.UpdateAddress)             // 编辑地址
		users.DELETE("/addresses/:id", userHandler.DeleteAddress)          // 删除地址

		// 头像管理
		users.GET("/avatar", userHandler.GetAvatar)                        // 头像信息
		users.POST("/avatar/upload", userHandler.UploadAvatar)            // 上传头像
		users.DELETE("/avatar", userHandler.RemoveAvatar)                 // 移除头像

		// 其他功能
		users.POST("/check-username", userHandler.CheckUsername)          // 检查用户名可用性
		users.GET("/downloadable-products", userHandler.GetDownloadableProducts) // 可下载产品
		users.POST("/external-association/remove", userHandler.RemoveExternalAssociation) // 移除外部关联
	}
}