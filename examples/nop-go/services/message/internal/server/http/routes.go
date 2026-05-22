package http

import (
	"nop-go/services/message/internal/server/http/handler"
	"nop-go/services/message/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册消息服务 HTTP 路由。
//
// 路由设计：
// - GET    /api/v1/message-templates          — 消息模板列表
// - GET    /api/v1/message-templates/:id      — 消息模板详情
// - POST   /api/v1/message-templates          — 创建消息模板
// - PUT    /api/v1/message-templates/:id      — 更新消息模板
// - DELETE /api/v1/message-templates/:id      — 删除消息模板
// - POST   /api/v1/message-templates/:id/test — 测试消息模板
// - POST   /api/v1/message-templates/:id/copy — 复制消息模板
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	messageHandler := handler.NewMessageHandler(services.Message)

	api := r.Group("/api/v1/message-templates")
	{
		api.GET("", messageHandler.List)              // 消息模板列表
		api.GET("/:id", messageHandler.GetByID)       // 消息模板详情
		api.POST("", messageHandler.Create)           // 创建消息模板
		api.PUT("/:id", messageHandler.Update)        // 更新消息模板
		api.DELETE("/:id", messageHandler.Delete)     // 删除消息模板
		api.POST("/:id/test", messageHandler.Test)    // 测试消息模板
		api.POST("/:id/copy", messageHandler.Copy)    // 复制消息模板
	}
}
