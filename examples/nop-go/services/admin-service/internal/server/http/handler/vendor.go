// Package handler 供应商模块处理器 —— 处理供应商管理的HTTP请求
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/service"
)

// VendorHandler 供应商处理器 —— 处理供应商管理相关的HTTP请求
type VendorHandler struct {
	svc *service.VendorService
}

// NewVendorHandler 创建供应商处理器
func NewVendorHandler(svc *service.VendorService) *VendorHandler {
	return &VendorHandler{svc: svc}
}

// Create 创建供应商
// POST /api/v1/vendors
func (h *VendorHandler) Create(c gorp.Context) {
	var req request.CreateVendorRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	v := &biz.Vendor{
		Name:        req.Name,
		Code:        req.Code,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Category:    req.Category,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		Status:      req.Status,
	}
	if err := h.svc.CreateVendor(c.Context(), v); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, v)
}

// Get 获取供应商详情
// GET /api/v1/vendors/:id
func (h *VendorHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	v, err := h.svc.GetVendor(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, v)
}

// List 获取供应商列表
// GET /api/v1/vendors
func (h *VendorHandler) List(c gorp.Context) {
	category := c.Query("category")
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	vendors, total, err := h.svc.ListVendors(c.Context(), category, status, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": vendors,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Update 更新供应商
// PUT /api/v1/vendors/:id
func (h *VendorHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	var req request.UpdateVendorRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	v := &biz.Vendor{
		ID:          uint(id),
		Name:        req.Name,
		Contact:     req.Contact,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		Category:    req.Category,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		Status:      req.Status,
	}
	if err := h.svc.UpdateVendor(c.Context(), v); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, v)
}

// Delete 删除供应商
// DELETE /api/v1/vendors/:id
func (h *VendorHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID")
		return
	}

	if err := h.svc.DeleteVendor(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}
