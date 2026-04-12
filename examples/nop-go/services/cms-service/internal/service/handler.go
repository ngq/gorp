// Package service CMS服务HTTP层
package service

import (
	"net/http"
	"strconv"

	"nop-go/services/cms-service/internal/biz"
	"nop-go/services/cms-service/internal/models"

	"github.com/gin-gonic/gin"
)

type CMSService struct {
	blogUC  *biz.BlogUseCase
	newsUC  *biz.NewsUseCase
	topicUC *biz.TopicUseCase
	forumUC *biz.ForumUseCase
	menuUC  *biz.MenuUseCase
	pollUC  *biz.PollUseCase
	htmlUC  *biz.HtmlBodyUseCase
}

func NewCMSService(
	blogUC *biz.BlogUseCase,
	newsUC *biz.NewsUseCase,
	topicUC *biz.TopicUseCase,
	forumUC *biz.ForumUseCase,
	menuUC *biz.MenuUseCase,
	pollUC *biz.PollUseCase,
	htmlUC *biz.HtmlBodyUseCase,
) *CMSService {
	return &CMSService{
		blogUC:  blogUC,
		newsUC:  newsUC,
		topicUC: topicUC,
		forumUC: forumUC,
		menuUC:  menuUC,
		pollUC:  pollUC,
		htmlUC:  htmlUC,
	}
}

func (s *CMSService) RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		blog := api.Group("/blog")
		{
			blog.GET("/posts", s.ListBlogPosts)
			blog.POST("/posts", s.CreateBlogPost)
			blog.GET("/posts/:id", s.GetBlogPost)
			blog.PUT("/posts/:id", s.UpdateBlogPost)
			blog.DELETE("/posts/:id", s.DeleteBlogPost)
			blog.POST("/posts/:id/publish", s.PublishBlogPost)
			blog.GET("/categories", s.ListBlogCategories)
			blog.POST("/categories", s.CreateBlogCategory)
		}

		news := api.Group("/news")
		{
			news.GET("", s.ListNews)
			news.POST("", s.CreateNews)
			news.GET("/:id", s.GetNews)
			news.PUT("/:id", s.UpdateNews)
			news.DELETE("/:id", s.DeleteNews)
		}

		topics := api.Group("/topics")
		{
			topics.GET("", s.ListTopics)
			topics.POST("", s.CreateTopic)
			topics.GET("/:id", s.GetTopic)
			topics.PUT("/:id", s.UpdateTopic)
			topics.DELETE("/:id", s.DeleteTopic)
		}

		forums := api.Group("/forums")
		{
			forums.GET("", s.ListForums)
			forums.POST("", s.CreateForum)
		}

		// 菜单管理
		menus := api.Group("/menus")
		{
			menus.GET("", s.ListMenus)
			menus.POST("", s.CreateMenu)
			menus.GET("/:id", s.GetMenu)
			menus.PUT("/:id", s.UpdateMenu)
			menus.DELETE("/:id", s.DeleteMenu)
			menus.GET("/:id/items", s.GetMenuItems)
			menus.POST("/items", s.CreateMenuItem)
			menus.PUT("/items/:id", s.UpdateMenuItem)
			menus.DELETE("/items/:id", s.DeleteMenuItem)
		}

		// 投票管理
		polls := api.Group("/polls")
		{
			polls.GET("", s.ListPolls)
			polls.GET("/homepage", s.ListHomepagePolls)
			polls.POST("", s.CreatePoll)
			polls.GET("/:id", s.GetPoll)
			polls.DELETE("/:id", s.DeletePoll)
			polls.POST("/:id/vote", s.Vote)
			polls.GET("/:id/result", s.GetPollResult)
		}

		// HTML内容块
		html := api.Group("/html")
		{
			html.GET("", s.ListHtmlBodies)
			html.POST("", s.CreateHtmlBody)
			html.GET("/:id", s.GetHtmlBody)
			html.GET("/name/:name", s.GetHtmlBodyByName)
			html.PUT("/:id", s.UpdateHtmlBody)
			html.DELETE("/:id", s.DeleteHtmlBody)
		}
	}
}

func (s *CMSService) ListBlogPosts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	posts, total, err := s.blogUC.ListPublishedPosts(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]models.BlogPostResponse, len(posts))
	for i, p := range posts {
		items[i] = models.ToBlogPostResponse(p)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (s *CMSService) CreateBlogPost(c *gin.Context) {
	var post models.BlogPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.blogUC.CreatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, post)
}

func (s *CMSService) GetBlogPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	post, err := s.blogUC.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, models.ToBlogPostResponse(post))
}

func (s *CMSService) UpdateBlogPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var post models.BlogPost
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.ID = id
	if err := s.blogUC.UpdatePost(c.Request.Context(), &post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, post)
}

func (s *CMSService) DeleteBlogPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.blogUC.DeletePost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CMSService) PublishBlogPost(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.blogUC.PublishPost(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "post published"})
}

func (s *CMSService) ListBlogCategories(c *gin.Context) {
	list, err := s.blogUC.ListCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreateBlogCategory(c *gin.Context) {
	var category models.BlogCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.blogUC.CreateCategory(c.Request.Context(), &category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, category)
}

func (s *CMSService) ListNews(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	news, total, err := s.newsUC.ListNews(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": news, "total": total})
}

func (s *CMSService) CreateNews(c *gin.Context) {
	var news models.News
	if err := c.ShouldBindJSON(&news); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.newsUC.CreateNews(c.Request.Context(), &news); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, news)
}

func (s *CMSService) GetNews(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	news, err := s.newsUC.GetNews(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}

func (s *CMSService) UpdateNews(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var news models.News
	if err := c.ShouldBindJSON(&news); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	news.ID = id
	if err := s.newsUC.UpdateNews(c.Request.Context(), &news); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, news)
}

func (s *CMSService) DeleteNews(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.newsUC.DeleteNews(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CMSService) ListTopics(c *gin.Context) {
	list, err := s.topicUC.ListTopics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreateTopic(c *gin.Context) {
	var topic models.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.topicUC.CreateTopic(c.Request.Context(), &topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, topic)
}

func (s *CMSService) GetTopic(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	topic, err := s.topicUC.GetTopic(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topic)
}

func (s *CMSService) UpdateTopic(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var topic models.Topic
	if err := c.ShouldBindJSON(&topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	topic.ID = id
	if err := s.topicUC.UpdateTopic(c.Request.Context(), &topic); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, topic)
}

func (s *CMSService) DeleteTopic(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.topicUC.DeleteTopic(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CMSService) ListForums(c *gin.Context) {
	list, err := s.forumUC.ListForums(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreateForum(c *gin.Context) {
	var forum models.Forum
	if err := c.ShouldBindJSON(&forum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.forumUC.CreateForum(c.Request.Context(), &forum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, forum)
}

// ========== 菜单接口 ==========

func (s *CMSService) ListMenus(c *gin.Context) {
	list, err := s.menuUC.ListMenus(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreateMenu(c *gin.Context) {
	var req models.MenuCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	menu, err := s.menuUC.CreateMenu(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, menu)
}

func (s *CMSService) GetMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	menu, err := s.menuUC.GetMenu(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, menu)
}

func (s *CMSService) UpdateMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.MenuUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	menu, err := s.menuUC.UpdateMenu(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, menu)
}

func (s *CMSService) DeleteMenu(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.menuUC.DeleteMenu(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CMSService) GetMenuItems(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	items, err := s.menuUC.GetMenuItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (s *CMSService) CreateMenuItem(c *gin.Context) {
	var req models.MenuItemCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := s.menuUC.CreateMenuItem(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (s *CMSService) UpdateMenuItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.MenuItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := s.menuUC.UpdateMenuItem(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (s *CMSService) DeleteMenuItem(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.menuUC.DeleteMenuItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ========== 投票接口 ==========

func (s *CMSService) ListPolls(c *gin.Context) {
	list, err := s.pollUC.ListPolls(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) ListHomepagePolls(c *gin.Context) {
	list, err := s.pollUC.ListHomepagePolls(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreatePoll(c *gin.Context) {
	var req models.PollCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	poll, err := s.pollUC.CreatePoll(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, poll)
}

func (s *CMSService) GetPoll(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	poll, err := s.pollUC.GetPoll(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, poll)
}

func (s *CMSService) DeletePoll(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.pollUC.DeletePoll(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *CMSService) Vote(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.PollVoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.PollID = id
	req.IPAddress = c.ClientIP()
	if err := s.pollUC.Vote(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vote recorded"})
}

func (s *CMSService) GetPollResult(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	result, err := s.pollUC.GetPollResult(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// ========== HTML内容块接口 ==========

func (s *CMSService) ListHtmlBodies(c *gin.Context) {
	list, err := s.htmlUC.ListHtmlBodies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (s *CMSService) CreateHtmlBody(c *gin.Context) {
	var req models.HtmlBodyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	html, err := s.htmlUC.CreateHtmlBody(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, html)
}

func (s *CMSService) GetHtmlBody(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	html, err := s.htmlUC.GetHtmlBody(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, html)
}

func (s *CMSService) GetHtmlBodyByName(c *gin.Context) {
	name := c.Param("name")
	html, err := s.htmlUC.GetHtmlBodyByName(c.Request.Context(), name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, html)
}

func (s *CMSService) UpdateHtmlBody(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req models.HtmlBodyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	html, err := s.htmlUC.UpdateHtmlBody(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, html)
}

func (s *CMSService) DeleteHtmlBody(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := s.htmlUC.DeleteHtmlBody(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}