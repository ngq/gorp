// Package handler 优惠模块处理器 —— 处理优惠与使用记录的HTTP请求
//
// 重要：本文件是从原 discount 独立服务合并而来。原 biz 层直接返回 response DTO，
// 现已重构为 biz 层返回领域实体，handler 层负责实体→响应 DTO 的转换。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/server/http/response"
	"nop-go/services/admin-service/internal/service"
)

// DiscountHandler 优惠处理器 —— 处理优惠管理相关的HTTP请求
type DiscountHandler struct {
	svc *service.DiscountService
}

// NewDiscountHandler 创建优惠处理器
func NewDiscountHandler(svc *service.DiscountService) *DiscountHandler {
	return &DiscountHandler{svc: svc}
}

// Create 创建优惠
// POST /api/v1/discounts
func (h *DiscountHandler) Create(c gorp.Context) {
	var req request.CreateDiscountRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	d := &biz.Discount{
		Name:         req.Name,
		Code:         req.Code,
		Type:         req.Type,
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		MaxDiscount:  req.MaxDiscount,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		TotalQuota:   req.TotalQuota,
		PerUserLimit: req.PerUserLimit,
		Status:       req.Status,
		Description:  req.Description,
	}
	if err := h.svc.CreateDiscount(c.Context(), d); err != nil {
		gorp.Error(c, err)
		return
	}

	// 领域实体 → 响应 DTO
	gorp.SuccessWithStatus(c, http.StatusCreated, response.NewDiscountResponse(d))
}

// Get 获取优惠详情
// GET /api/v1/discounts/:id
func (h *DiscountHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	d, err := h.svc.GetDiscount(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 领域实体 → 响应 DTO
	gorp.Success(c, response.NewDiscountResponse(d))
}

// List 获取优惠列表
// GET /api/v1/discounts
func (h *DiscountHandler) List(c gorp.Context) {
	discountType, _ := strconv.Atoi(c.DefaultQuery("type", "0"))
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	discounts, total, err := h.svc.ListDiscounts(c.Context(), discountType, status, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 领域实体列表 → 响应 DTO 列表
	items := response.NewDiscountResponseList(discounts)
	gorp.Success(c, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Update 更新优惠
// PUT /api/v1/discounts/:id
func (h *DiscountHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	var req request.UpdateDiscountRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	d := &biz.Discount{
		ID:           uint(id),
		Name:         req.Name,
		Type:         req.Type,
		Value:        req.Value,
		MinAmount:    req.MinAmount,
		MaxDiscount:  req.MaxDiscount,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		TotalQuota:   req.TotalQuota,
		PerUserLimit: req.PerUserLimit,
		Status:       req.Status,
		Description:  req.Description,
	}
	if err := h.svc.UpdateDiscount(c.Context(), d); err != nil {
		gorp.Error(c, err)
		return
	}

	// 领域实体 → 响应 DTO
	gorp.Success(c, response.NewDiscountResponse(d))
}

// Delete 删除优惠
// DELETE /api/v1/discounts/:id
func (h *DiscountHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.DeleteDiscount(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ListUsages 获取优惠使用记录列表
// GET /api/v1/discounts/:id/usages
func (h *DiscountHandler) ListUsages(c gorp.Context) {
	discountID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的优惠ID")
		return
	}
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	usages, total, err := h.svc.ListDiscountUsages(c.Context(), uint(discountID), uint(userID), page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 领域实体列表 → 响应 DTO 列表
	items := response.NewDiscountUsageResponseList(usages)
	gorp.Success(c, map[string]any{
		"items": items,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}
