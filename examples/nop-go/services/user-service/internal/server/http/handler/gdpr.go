// Package handler 定义 GDPR 相关的 HTTP 请求处理器。
// 本文件合并了原 gdpr 服务的处理器逻辑，使用 gorp.Context 抽象处理 HTTP 请求。
package handler

import (
	"net/http"
	"strconv"

	"nop-go/services/user-service/internal/biz"
	"nop-go/services/user-service/internal/server/http/request"
	"nop-go/services/user-service/internal/server/http/response"
	"nop-go/services/user-service/internal/service"

	gorp "github.com/ngq/gorp"
)

// GdprHandler GDPR 相关的 HTTP 请求处理器
type GdprHandler struct {
	gdpr *service.GdprService
}

// NewGdprHandler 创建 GDPR 处理器实例
func NewGdprHandler(gdpr *service.GdprService) *GdprHandler {
	return &GdprHandler{gdpr: gdpr}
}

// List 获取 GDPR 请求列表（管理端分页查询）。
// GET /api/v1/gdprs
func (h *GdprHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.gdpr.ListGdprs(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 将 service 层的 DTO 转换为 HTTP 响应格式
	respItems := make([]*response.GdprResponse, len(items))
	for i, item := range items {
		respItems[i] = &response.GdprResponse{
			ID:          item.ID,
			UserID:      item.UserID,
			RequestType: item.RequestType,
			Status:      item.Status,
			Reason:      item.Reason,
			ReviewedBy:  item.ReviewedBy,
			ReviewedAt:  item.ReviewedAt,
			CompletedAt: item.CompletedAt,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	gorp.Success(c, response.GdprListResponse{
		Data:  respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetByID 获取单个 GDPR 请求。
// GET /api/v1/gdprs/:id
func (h *GdprHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的 GDPR 请求 ID")
		return
	}

	dto, err := h.gdpr.GetGdpr(c.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, map[string]any{"error": "GDPR 请求不存在"})
		return
	}

	gorp.Success(c, response.GdprResponse{
		ID:          dto.ID,
		UserID:      dto.UserID,
		RequestType: dto.RequestType,
		Status:      dto.Status,
		Reason:      dto.Reason,
		ReviewedBy:  dto.ReviewedBy,
		ReviewedAt:  dto.ReviewedAt,
		CompletedAt: dto.CompletedAt,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	})
}

// Create 创建 GDPR 请求。
// POST /api/v1/gdprs
func (h *GdprHandler) Create(c gorp.Context) {
	var req request.CreateGdprRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	gdpr := &biz.Gdpr{
		UserID:      req.UserID,
		RequestType: req.RequestType,
		Reason:      req.Reason,
	}

	dto, err := h.gdpr.CreateGdpr(c.Context(), gdpr)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, response.GdprResponse{
		ID:          dto.ID,
		UserID:      dto.UserID,
		RequestType: dto.RequestType,
		Status:      dto.Status,
		Reason:      dto.Reason,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	})
}

// Update 更新 GDPR 请求。
// PUT /api/v1/gdprs/:id
func (h *GdprHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的 GDPR 请求 ID")
		return
	}

	var req request.UpdateGdprRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	gdpr := &biz.Gdpr{
		ID:          uint(id),
		RequestType: req.RequestType,
		Reason:      req.Reason,
	}

	dto, err := h.gdpr.UpdateGdpr(c.Context(), gdpr)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.GdprResponse{
		ID:          dto.ID,
		UserID:      dto.UserID,
		RequestType: dto.RequestType,
		Status:      dto.Status,
		Reason:      dto.Reason,
		ReviewedBy:  dto.ReviewedBy,
		ReviewedAt:  dto.ReviewedAt,
		CompletedAt: dto.CompletedAt,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	})
}

// Delete 删除 GDPR 请求。
// DELETE /api/v1/gdprs/:id
func (h *GdprHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的 GDPR 请求 ID")
		return
	}

	if err := h.gdpr.DeleteGdpr(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// Review 审核 GDPR 请求。
// PUT /api/v1/gdprs/:id/review
func (h *GdprHandler) Review(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的 GDPR 请求 ID")
		return
	}

	var req request.ReviewGdprRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	dto, err := h.gdpr.ReviewGdpr(c.Context(), uint(id), req.Status, req.ReviewedBy)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.GdprResponse{
		ID:          dto.ID,
		UserID:      dto.UserID,
		RequestType: dto.RequestType,
		Status:      dto.Status,
		Reason:      dto.Reason,
		ReviewedBy:  dto.ReviewedBy,
		ReviewedAt:  dto.ReviewedAt,
		CompletedAt: dto.CompletedAt,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	})
}

// Complete 完成 GDPR 请求。
// PUT /api/v1/gdprs/:id/complete
func (h *GdprHandler) Complete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的 GDPR 请求 ID")
		return
	}

	dto, err := h.gdpr.CompleteGdpr(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, response.GdprResponse{
		ID:          dto.ID,
		UserID:      dto.UserID,
		RequestType: dto.RequestType,
		Status:      dto.Status,
		Reason:      dto.Reason,
		ReviewedBy:  dto.ReviewedBy,
		ReviewedAt:  dto.ReviewedAt,
		CompletedAt: dto.CompletedAt,
		CreatedAt:   dto.CreatedAt,
		UpdatedAt:   dto.UpdatedAt,
	})
}