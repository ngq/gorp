// Package handler 插件模块处理器 —— 处理插件管理的HTTP请求
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/service"
)

// PluginHandler 插件处理器 —— 处理插件管理相关的HTTP请求
type PluginHandler struct {
	svc *service.PluginService
}

// NewPluginHandler 创建插件处理器
func NewPluginHandler(svc *service.PluginService) *PluginHandler {
	return &PluginHandler{svc: svc}
}

// List 获取插件列表
func (h *PluginHandler) List(c gorp.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	plugins, total, err := h.svc.ListPlugins(c.Context(), status, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": plugins,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取插件详情
func (h *PluginHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	plugin, err := h.svc.GetPlugin(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, plugin)
}

// Create 创建插件
func (h *PluginHandler) Create(c gorp.Context) {
	var req request.CreatePluginRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	p := &biz.Plugin{
		Name:        req.Name,
		Code:        req.Code,
		Version:     req.Version,
		Description: req.Description,
		Author:      req.Author,
		Config:      req.Config,
		Status:      req.Status,
		Sort:        req.Sort,
	}
	if err := h.svc.CreatePlugin(c.Context(), p); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, p)
}

// Update 更新插件
func (h *PluginHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdatePluginRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	p := &biz.Plugin{
		ID:          uint(id),
		Name:        req.Name,
		Version:     req.Version,
		Description: req.Description,
		Author:      req.Author,
		Config:      req.Config,
		Status:      req.Status,
		Sort:        req.Sort,
	}
	if err := h.svc.UpdatePlugin(c.Context(), p); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, p)
}

// Delete 删除插件
func (h *PluginHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.svc.DeletePlugin(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}