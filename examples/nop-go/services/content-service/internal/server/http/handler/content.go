package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/content-service/internal/server/http/request"
	"nop-go/services/content-service/internal/server/http/response"
	"nop-go/services/content-service/internal/service"
)

// ContentHandler 内容服务 HTTP 处理器
type ContentHandler struct {
	content *service.ContentService
}

// NewContentHandler 创建内容服务处理器
func NewContentHandler(content *service.ContentService) *ContentHandler {
	return &ContentHandler{content: content}
}

// ==================== 博客 ====================

// ListBlogs 博客列表
func (h *ContentHandler) ListBlogs(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListBlogs(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.BlogList{Items: items, Total: total, Page: page, Size: size})
}

// CreateBlog 创建博客
func (h *ContentHandler) CreateBlog(c gorp.Context) {
	var req request.CreateBlog
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	blog, err := h.content.CreateBlog(c.Context(), service.BlogRequest{
		Title: req.Title, Content: req.Content, Author: req.Author,
		Status: req.Status, Tags: req.Tags,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, blog)
}

// GetBlog 获取博客详情
func (h *ContentHandler) GetBlog(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	blog, err := h.content.GetBlog(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, blog)
}

// UpdateBlog 更新博客
func (h *ContentHandler) UpdateBlog(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdateBlog
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	blog, err := h.content.UpdateBlog(c.Context(), id, service.BlogRequest{
		Title: req.Title, Content: req.Content, Author: req.Author,
		Status: req.Status, Tags: req.Tags,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, blog)
}

// DeleteBlog 删除博客
func (h *ContentHandler) DeleteBlog(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteBlog(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 新闻 ====================

// ListNews 新闻列表
func (h *ContentHandler) ListNews(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListNews(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.NewsList{Items: items, Total: total, Page: page, Size: size})
}

// CreateNews 创建新闻
func (h *ContentHandler) CreateNews(c gorp.Context) {
	var req request.CreateNews
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	news, err := h.content.CreateNews(c.Context(), service.NewsRequest{
		Title: req.Title, Content: req.Content, Source: req.Source,
		Category: req.Category, Priority: req.Priority, Status: req.Status,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, news)
}

// GetNews 获取新闻详情
func (h *ContentHandler) GetNews(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	news, err := h.content.GetNews(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, news)
}

// UpdateNews 更新新闻
func (h *ContentHandler) UpdateNews(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdateNews
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	news, err := h.content.UpdateNews(c.Context(), id, service.NewsRequest{
		Title: req.Title, Content: req.Content, Source: req.Source,
		Category: req.Category, Priority: req.Priority, Status: req.Status,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, news)
}

// DeleteNews 删除新闻
func (h *ContentHandler) DeleteNews(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteNews(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 话题 ====================

// ListTopics 话题列表
func (h *ContentHandler) ListTopics(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListTopics(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.TopicList{Items: items, Total: total, Page: page, Size: size})
}

// CreateTopic 创建话题
func (h *ContentHandler) CreateTopic(c gorp.Context) {
	var req request.CreateTopic
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	topic, err := h.content.CreateTopic(c.Context(), service.TopicRequest{
		Title: req.Title, Description: req.Description,
		CoverImage: req.CoverImage, SortOrder: req.SortOrder, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, topic)
}

// GetTopic 获取话题详情
func (h *ContentHandler) GetTopic(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	topic, err := h.content.GetTopic(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, topic)
}

// UpdateTopic 更新话题
func (h *ContentHandler) UpdateTopic(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdateTopic
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	topic, err := h.content.UpdateTopic(c.Context(), id, service.TopicRequest{
		Title: req.Title, Description: req.Description,
		CoverImage: req.CoverImage, SortOrder: req.SortOrder, IsActive: req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, topic)
}

// DeleteTopic 删除话题
func (h *ContentHandler) DeleteTopic(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteTopic(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 投票 ====================

// ListPolls 投票列表
func (h *ContentHandler) ListPolls(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListPolls(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.PollList{Items: items, Total: total, Page: page, Size: size})
}

// CreatePoll 创建投票
func (h *ContentHandler) CreatePoll(c gorp.Context) {
	var req request.CreatePoll
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	poll, err := h.content.CreatePoll(c.Context(), service.PollRequest{
		Title: req.Title, Question: req.Question, Options: req.Options,
		IsActive: req.IsActive, EndTime: req.EndTime,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, poll)
}

// GetPoll 获取投票详情
func (h *ContentHandler) GetPoll(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	poll, err := h.content.GetPoll(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, poll)
}

// UpdatePoll 更新投票
func (h *ContentHandler) UpdatePoll(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdatePoll
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	poll, err := h.content.UpdatePoll(c.Context(), id, service.PollRequest{
		Title: req.Title, Question: req.Question, Options: req.Options,
		IsActive: req.IsActive, EndTime: req.EndTime,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, poll)
}

// DeletePoll 删除投票
func (h *ContentHandler) DeletePoll(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeletePoll(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}