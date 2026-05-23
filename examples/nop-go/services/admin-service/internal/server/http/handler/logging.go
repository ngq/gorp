// Package handler 日志模块处理器 —— 处理活动日志与系统日志的HTTP请求
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/service"
)

// ==================== 活动日志处理器 ====================

// ActivityLogHandler 活动日志处理器
type ActivityLogHandler struct {
	svc *service.LoggingService
}

// NewActivityLogHandler 创建活动日志处理器
func NewActivityLogHandler(svc *service.LoggingService) *ActivityLogHandler {
	return &ActivityLogHandler{svc: svc}
}

// List 获取活动日志列表
func (h *ActivityLogHandler) List(c gorp.Context) {
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	action := c.DefaultQuery("action", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logs, total, err := h.svc.ListActivityLogs(c.Context(), uint(userID), action, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": logs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取活动日志详情
func (h *ActivityLogHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	log, err := h.svc.GetActivityLog(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, log)
}

// Create 创建活动日志
func (h *ActivityLogHandler) Create(c gorp.Context) {
	var req request.CreateActivityLogRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	log := &biz.ActivityLog{
		UserID:    req.UserID,
		UserName:  req.UserName,
		Action:    req.Action,
		Resource:  req.Resource,
		IP:        req.IP,
		UserAgent: req.UserAgent,
		Detail:    req.Detail,
		Status:    req.Status,
	}
	if err := h.svc.CreateActivityLog(c.Context(), log); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, log)
}

// ==================== 系统日志处理器 ====================

// SystemLogHandler 系统日志处理器
type SystemLogHandler struct {
	svc *service.LoggingService
}

// NewSystemLogHandler 创建系统日志处理器
func NewSystemLogHandler(svc *service.LoggingService) *SystemLogHandler {
	return &SystemLogHandler{svc: svc}
}

// List 获取系统日志列表
func (h *SystemLogHandler) List(c gorp.Context) {
	level := c.DefaultQuery("level", "")
	module := c.DefaultQuery("module", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	logs, total, err := h.svc.ListSystemLogs(c.Context(), level, module, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": logs,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取系统日志详情
func (h *SystemLogHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	log, err := h.svc.GetSystemLog(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, log)
}

// Create 创建系统日志
func (h *SystemLogHandler) Create(c gorp.Context) {
	var req request.CreateSystemLogRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	log := &biz.SystemLog{
		Level:    req.Level,
		Module:   req.Module,
		Message:  req.Message,
		Stack:    req.Stack,
		Hostname: req.Hostname,
	}
	if err := h.svc.CreateSystemLog(c.Context(), log); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, log)
}