package handler

import (
	"net/http"
	"strconv"
	gorp "github.com/ngq/gorp"
	"nop-go/services/vendorsvc/internal/server/http/request"
	"nop-go/services/vendorsvc/internal/server/http/response"
	"nop-go/services/vendorsvc/internal/service"
)

// VendorHandler 供应商服务 HTTP 处理器。
type VendorHandler struct {
	vendor *service.VendorService
}

// NewVendorHandler 创建供应商服务处理器。
func NewVendorHandler(vendor *service.VendorService) *VendorHandler {
	return &VendorHandler{vendor: vendor}
}

// List 供应商列表。
func (h *VendorHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.vendor.List(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.Vendor, len(items))
	for i, item := range items {
		respItems[i] = response.Vendor{
			ID:           item.ID,
			Name:         item.Name,
			Email:        item.Email,
			Active:       item.Active,
			DisplayOrder: item.DisplayOrder,
		}
	}

	gorp.Success(c, response.VendorList{
		Items: respItems,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建供应商。
func (h *VendorHandler) Create(c gorp.Context) {
	var req request.CreateVendor
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	vendor, err := h.vendor.Create(c.Context(), service.CreateVendorRequest{
		Name:         req.Name,
		Email:        req.Email,
		Description:  req.Description,
		Active:       req.Active,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, vendor)
}

// Update 更新供应商。
func (h *VendorHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateVendor
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	vendor, err := h.vendor.Update(c.Context(), uint(id), service.UpdateVendorRequest{
		Name:         req.Name,
		Email:        req.Email,
		Description:  req.Description,
		Active:       req.Active,
		DisplayOrder: req.DisplayOrder,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, vendor)
}

// Delete 删除供应商。
func (h *VendorHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.vendor.Delete(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// GetApply 查询供应商申请状态。
func (h *VendorHandler) GetApply(c gorp.Context) {
	// 从查询参数获取申请ID
	applyIDStr := c.Query("apply_id")
	if applyIDStr == "" {
		gorp.BadRequest(c, "缺少申请ID参数")
		return
	}

	applyID, err := strconv.ParseUint(applyIDStr, 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的申请ID参数")
		return
	}

	apply, err := h.vendor.GetApply(c.Context(), uint(applyID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, apply)
}

// SubmitApply 提交供应商申请。
func (h *VendorHandler) SubmitApply(c gorp.Context) {
	var req request.VendorApply
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	apply, err := h.vendor.SubmitApply(c.Context(), service.VendorApplyRequest{
		Name:        req.Name,
		Email:       req.Email,
		Description: req.Description,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, apply)
}