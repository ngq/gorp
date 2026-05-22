package http

import (
	"nop-go/services/localization/internal/server/http/handler"
	"nop-go/services/localization/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册本地化服务 HTTP 路由。
//
// 路由设计：
// - GET    /api/v1/languages                  — 语言列表
// - POST   /api/v1/languages                  — 创建语言
// - PUT    /api/v1/languages/:id              — 更新语言
// - DELETE /api/v1/languages/:id              — 删除语言
// - GET    /api/v1/languages/:id/resources     — 本地化资源列表
// - POST   /api/v1/languages/:id/resources     — 添加本地化资源
// - PUT    /api/v1/languages/resources/:id    — 更新本地化资源
// - DELETE /api/v1/languages/resources/:id    — 删除本地化资源
// - POST   /api/v1/languages/:id/export       — 导出语言资源
// - POST   /api/v1/languages/:id/import       — 导入语言资源
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	locHandler := handler.NewLocalizationHandler(services.Localization)

	// 语言相关路由
	lang := r.Group("/api/v1/languages")
	{
		lang.GET("", locHandler.ListLanguages)              // 语言列表
		lang.POST("", locHandler.CreateLanguage)            // 创建语言
		lang.PUT("/:id", locHandler.UpdateLanguage)         // 更新语言
		lang.DELETE("/:id", locHandler.DeleteLanguage)      // 删除语言
		lang.GET("/:id/resources", locHandler.ListResources)       // 本地化资源列表
		lang.POST("/:id/resources", locHandler.AddResource)        // 添加本地化资源
		lang.PUT("/resources/:id", locHandler.UpdateResource)     // 更新本地化资源
		lang.DELETE("/resources/:id", locHandler.DeleteResource)  // 删除本地化资源
		lang.POST("/:id/export", locHandler.ExportResources)      // 导出语言资源
		lang.POST("/:id/import", locHandler.ImportResources)      // 导入语言资源
	}
}