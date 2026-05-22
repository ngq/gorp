package http

import (
	"nop-go/services/media/internal/server/http/handler"
	"nop-go/services/media/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册媒体服务 HTTP 路由。
//
// 路由设计：
// - POST /api/v1/media/upload   — 异步上传图片
// - GET  /api/v1/media/:id      — 获取图片
// - DELETE /api/v1/media/:id    — 删除图片
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	mediaHandler := handler.NewMediaHandler(services.Media)

	api := r.Group("/api/v1/media")
	{
		api.POST("/upload", mediaHandler.Upload)   // 异步上传图片
		api.GET("/:id", mediaHandler.GetByID)      // 获取图片
		api.DELETE("/:id", mediaHandler.Delete)    // 删除图片
	}
}
