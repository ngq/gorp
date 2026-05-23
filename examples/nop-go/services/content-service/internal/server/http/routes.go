package http

import (
	"nop-go/services/content-service/internal/server/http/handler"
	"nop-go/services/content-service/internal/service"

	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册内容聚合服务 HTTP 路由。
//
// 合并路由设计：
// - 博客：GET/POST/PUT/DELETE /api/v1/blog
// - 新闻：GET/POST/PUT/DELETE /api/v1/news
// - 话题：GET/POST/PUT/DELETE /api/v1/topics
// - 投票：GET/POST/PUT/DELETE /api/v1/polls
// - 语言：GET/POST/PUT/DELETE /api/v1/languages + 子资源
// - 推广：GET/POST/PUT/DELETE /api/v1/affiliates + 子资源
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	contentHandler := handler.NewContentHandler(services.Content)
	locHandler := handler.NewLocalizationHandler(services.Localization)
	affHandler := handler.NewAffiliateHandler(services.Affiliate)

	// ---- 博客路由 ----
	blog := r.Group("/api/v1/blog")
	{
		blog.GET("", contentHandler.ListBlogs)
		blog.POST("", contentHandler.CreateBlog)
		blog.GET("/:id", contentHandler.GetBlog)
		blog.PUT("/:id", contentHandler.UpdateBlog)
		blog.DELETE("/:id", contentHandler.DeleteBlog)
	}

	// ---- 新闻路由 ----
	news := r.Group("/api/v1/news")
	{
		news.GET("", contentHandler.ListNews)
		news.POST("", contentHandler.CreateNews)
		news.GET("/:id", contentHandler.GetNews)
		news.PUT("/:id", contentHandler.UpdateNews)
		news.DELETE("/:id", contentHandler.DeleteNews)
	}

	// ---- 话题路由 ----
	topics := r.Group("/api/v1/topics")
	{
		topics.GET("", contentHandler.ListTopics)
		topics.POST("", contentHandler.CreateTopic)
		topics.GET("/:id", contentHandler.GetTopic)
		topics.PUT("/:id", contentHandler.UpdateTopic)
		topics.DELETE("/:id", contentHandler.DeleteTopic)
	}

	// ---- 投票路由 ----
	polls := r.Group("/api/v1/polls")
	{
		polls.GET("", contentHandler.ListPolls)
		polls.POST("", contentHandler.CreatePoll)
		polls.GET("/:id", contentHandler.GetPoll)
		polls.PUT("/:id", contentHandler.UpdatePoll)
		polls.DELETE("/:id", contentHandler.DeletePoll)
	}

	// ---- 语言路由 ----
	languages := r.Group("/api/v1/languages")
	{
		languages.GET("", locHandler.ListLanguages)
		languages.POST("", locHandler.CreateLanguage)
		languages.GET("/:id", locHandler.GetLanguage)
		languages.PUT("/:id", locHandler.UpdateLanguage)
		languages.DELETE("/:id", locHandler.DeleteLanguage)
		// 语言下的本地化资源
		languages.GET("/:id/resources", locHandler.ListLocaleResources)
		languages.POST("/:id/resources", locHandler.CreateLocaleResource)
		languages.GET("/resources/:resId", locHandler.GetLocaleResource)
		languages.PUT("/resources/:resId", locHandler.UpdateLocaleResource)
		languages.DELETE("/resources/:resId", locHandler.DeleteLocaleResource)
	}

	// ---- 推广合作方路由 ----
	affiliates := r.Group("/api/v1/affiliates")
	{
		affiliates.GET("", affHandler.ListAffiliates)
		affiliates.POST("", affHandler.CreateAffiliate)
		affiliates.GET("/:id", affHandler.GetAffiliate)
		affiliates.PUT("/:id", affHandler.UpdateAffiliate)
		affiliates.DELETE("/:id", affHandler.DeleteAffiliate)
		// 推广合作方关联子资源
		affiliates.GET("/:id/orders", affHandler.ListOrders)
		affiliates.GET("/:id/customers", affHandler.ListCustomers)
	}
}