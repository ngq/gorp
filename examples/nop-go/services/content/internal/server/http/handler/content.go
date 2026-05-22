package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/content/internal/server/http/request"
	"nop-go/services/content/internal/server/http/response"
	"nop-go/services/content/internal/service"
)

// ContentHandler 内容服务 HTTP 处理器。
type ContentHandler struct {
	content *service.ContentService
}

// NewContentHandler 创建内容服务处理器。
func NewContentHandler(content *service.ContentService) *ContentHandler {
	return &ContentHandler{content: content}
}

// ==================== 博客 ====================

// ListBlog 博客列表。
func (h *ContentHandler) ListBlog(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListBlog(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.BlogList{Items: items, Total: total, Page: page, Size: size})
}

// CreateBlog 创建博客。
func (h *ContentHandler) CreateBlog(c gorp.Context) {
	var req request.CreateBlog
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	blog, err := h.content.CreateBlog(c.Context(), service.CreateBlogRequest{
		Title: req.Title, Body: req.Body, Tags: req.Tags, AllowComments: req.AllowComments,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, blog)
}

// UpdateBlog 更新博客。
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
	blog, err := h.content.UpdateBlog(c.Context(), uint(id), service.UpdateBlogRequest{
		Title: req.Title, Body: req.Body, Tags: req.Tags, AllowComments: req.AllowComments,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, blog)
}

// DeleteBlog 删除博客。
func (h *ContentHandler) DeleteBlog(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteBlog(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 新闻 ====================

// ListNews 新闻列表。
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

// CreateNews 创建新闻。
func (h *ContentHandler) CreateNews(c gorp.Context) {
	var req request.CreateNews
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	news, err := h.content.CreateNews(c.Context(), service.CreateNewsRequest{
		Title: req.Title, Body: req.Body, AllowComments: req.AllowComments,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, news)
}

// UpdateNews 更新新闻。
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
	news, err := h.content.UpdateNews(c.Context(), uint(id), service.UpdateNewsRequest{
		Title: req.Title, Body: req.Body, AllowComments: req.AllowComments,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, news)
}

// DeleteNews 删除新闻。
func (h *ContentHandler) DeleteNews(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteNews(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 页面 ====================

// ListTopic 页面列表。
func (h *ContentHandler) ListTopic(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListTopic(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.TopicList{Items: items, Total: total, Page: page, Size: size})
}

// CreateTopic 创建页面。
func (h *ContentHandler) CreateTopic(c gorp.Context) {
	var req request.CreateTopic
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	topic, err := h.content.CreateTopic(c.Context(), service.CreateTopicRequest{
		Title: req.Title, Body: req.Body, IsPublished: req.IsPublished,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, topic)
}

// UpdateTopic 更新页面。
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
	topic, err := h.content.UpdateTopic(c.Context(), uint(id), service.UpdateTopicRequest{
		Title: req.Title, Body: req.Body, IsPublished: req.IsPublished,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, topic)
}

// DeleteTopic 删除页面。
func (h *ContentHandler) DeleteTopic(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeleteTopic(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 投票 ====================

// ListPoll 投票列表。
func (h *ContentHandler) ListPoll(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.content.ListPoll(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.PollList{Items: items, Total: total, Page: page, Size: size})
}

// CreatePoll 创建投票。
func (h *ContentHandler) CreatePoll(c gorp.Context) {
	var req request.CreatePoll
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	poll, err := h.content.CreatePoll(c.Context(), service.CreatePollRequest{
		Name: req.Name, AllowSelectMultiple: req.AllowSelectMultiple,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, poll)
}

// UpdatePoll 更新投票。
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
	poll, err := h.content.UpdatePoll(c.Context(), uint(id), service.UpdatePollRequest{
		Name: req.Name, AllowSelectMultiple: req.AllowSelectMultiple,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, poll)
}

// DeletePoll 删除投票。
func (h *ContentHandler) DeletePoll(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.content.DeletePoll(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// VotePoll 投票。
func (h *ContentHandler) VotePoll(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.VotePoll
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	if err := h.content.VotePoll(c.Context(), uint(id), req.AnswerID); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, map[string]string{"message": "投票成功"})
}