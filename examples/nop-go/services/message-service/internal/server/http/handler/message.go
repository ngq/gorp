// Package handler HTTP 请求处理器。
//
// handler 层负责 HTTP 请求的解析、参数校验和响应返回，
// 是 HTTP 路由与 service 层之间的适配器。
package handler

import (
	"net/http"
	"strconv"
	"time"

	gorp "github.com/ngq/gorp"

	"nop-go/services/message-service/internal/server/http/request"
	"nop-go/services/message-service/internal/server/http/response"
	"nop-go/services/message-service/internal/service"
)

// toMessageTemplateResponse 将 service.MessageTemplateResponse 转换为 response.MessageTemplate。
//
// service 层返回的时间字段为格式化字符串，response 层需要 time.Time，
// 因此需要解析字符串还原为 time.Time。
func toMessageTemplateResponse(src *service.MessageTemplateResponse) response.MessageTemplate {
	return response.MessageTemplate{
		ID:           src.ID,
		Name:         src.Name,
		Subject:      src.Subject,
		Body:         src.Body,
		EmailAccount: src.EmailAccount,
		IsActive:     src.IsActive,
		CreatedAt:    parseMessageTime(src.CreatedAt),
		UpdatedAt:    parseMessageTime(src.UpdatedAt),
	}
}

// parseMessageTime 解析 service 层返回的时间字符串为 time.Time。
//
// service 层统一使用 "2006-01-02 15:04:05" 格式输出时间，
// 解析失败时返回零值。
func parseMessageTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// MessageHandler 消息服务 HTTP 处理器。
//
// 每个 handler 方法对应一个路由端点，负责：
// 1. 解析请求参数（路径参数、查询参数、请求体）
// 2. 调用 service 层执行业务逻辑
// 3. 构造并返回 HTTP 响应
type MessageHandler struct {
	message *service.MessageService
}

// NewMessageHandler 创建消息服务处理器。
func NewMessageHandler(message *service.MessageService) *MessageHandler {
	return &MessageHandler{message: message}
}

// List 消息模板列表。
//
// 支持分页查询，通过 page/size 查询参数控制。
func (h *MessageHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.message.List(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 将 service 层返回的 MessageTemplateResponse 转换为 response.MessageTemplate
	respItems := make([]response.MessageTemplate, len(items))
	for i, item := range items {
		respItems[i] = toMessageTemplateResponse(&item)
	}

	gorp.Success(c, response.MessageTemplateList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetByID 消息模板详情。
//
// 通过路径参数 :id 获取指定模板。
func (h *MessageHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	tpl, err := h.message.GetByID(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, tpl)
}

// Create 创建消息模板。
//
// 从请求体解析创建参数，调用 service 层创建模板。
func (h *MessageHandler) Create(c gorp.Context) {
	var req request.CreateMessageTemplate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	tpl, err := h.message.Create(c.Context(), service.CreateMessageTemplateRequest{
		Name:         req.Name,
		Subject:      req.Subject,
		Body:         req.Body,
		EmailAccount: req.EmailAccount,
		IsActive:     req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, tpl)
}

// Update 更新消息模板。
//
// 通过路径参数 :id 指定模板，请求体包含更新字段。
func (h *MessageHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateMessageTemplate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	tpl, err := h.message.Update(c.Context(), uint(id), service.UpdateMessageTemplateRequest{
		Name:         req.Name,
		Subject:      req.Subject,
		Body:         req.Body,
		EmailAccount: req.EmailAccount,
		IsActive:     req.IsActive,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, tpl)
}

// Delete 删除消息模板。
//
// 通过路径参数 :id 指定要删除的模板。
func (h *MessageHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.message.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// Test 测试消息模板，发送测试邮件。
//
// 通过路径参数 :id 指定模板，请求体包含收件邮箱。
func (h *MessageHandler) Test(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.TestMessageTemplate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	if err := h.message.Test(c.Context(), uint(id), req.ToEmail); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]string{"message": "测试消息已发送"})
}

// Copy 复制消息模板。
//
// 通过路径参数 :id 指定源模板，创建副本。
func (h *MessageHandler) Copy(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	tpl, err := h.message.Copy(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, tpl)
}
