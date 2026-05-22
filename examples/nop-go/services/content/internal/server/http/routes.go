package http

import (
	"nop-go/services/content/internal/server/http/handler"
	"nop-go/services/content/internal/service"
	gorp "github.com/ngq/gorp"
)

// RegisterRoutes 注册内容服务 HTTP 路由。
//
// 路由设计：
// 博客：GET/POST/PUT/DELETE /api/v1/blog
// 新闻：GET/POST/PUT/DELETE /api/v1/news
// 页面：GET/POST/PUT/DELETE /api/v1/topics
// 投票：GET/POST/PUT/DELETE /api/v1/polls + POST /api/v1/polls/:id/vote
func RegisterRoutes(
	r gorp.Router,
	services *service.Services,
) {
	contentHandler := handler.NewContentHandler(services.Content)

	// 博客路由
	blog := r.Group("/api/v1/blog")
	{
		blog.GET("", contentHandler.ListBlog)
		blog.POST("", contentHandler.CreateBlog)
		blog.PUT("/:id", contentHandler.UpdateBlog)
		blog.DELETE("/:id", contentHandler.DeleteBlog)
	}

	// 新闻路由
	news := r.Group("/api/v1/news")
	{
		news.GET("", contentHandler.ListNews)
		news.POST("", contentHandler.CreateNews)
		news.PUT("/:id", contentHandler.UpdateNews)
		news.DELETE("/:id", contentHandler.DeleteNews)
	}

	// 页面路由
	topics := r.Group("/api/v1/topics")
	{
		topics.GET("", contentHandler.ListTopic)
		topics.POST("", contentHandler.CreateTopic)
		topics.PUT("/:id", contentHandler.UpdateTopic)
		topics.DELETE("/:id", contentHandler.DeleteTopic)
	}

	// 投票路由
	polls := r.Group("/api/v1/polls")
	{
		polls.GET("", contentHandler.ListPoll)
		polls.POST("", contentHandler.CreatePoll)
		polls.PUT("/:id", contentHandler.UpdatePoll)
		polls.DELETE("/:id", contentHandler.DeletePoll)
		polls.POST("/:id/vote", contentHandler.VotePoll)
	}
}