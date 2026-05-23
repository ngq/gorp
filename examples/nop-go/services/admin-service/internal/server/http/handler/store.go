// Package handler 门店模块处理器 —— 处理门店管理的HTTP请求
package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"

	"nop-go/services/admin-service/internal/biz"
	"nop-go/services/admin-service/internal/server/http/request"
	"nop-go/services/admin-service/internal/service"
)

// StoreHandler 门店处理器 —— 处理门店管理相关的HTTP请求
type StoreHandler struct {
	svc *service.StoreService
}

// NewStoreHandler 创建门店处理器
func NewStoreHandler(svc *service.StoreService) *StoreHandler {
	return &StoreHandler{svc: svc}
}

// List 获取门店列表
func (h *StoreHandler) List(c gorp.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "-1"))
	region := c.DefaultQuery("region", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	stores, total, err := h.svc.ListStores(c.Context(), status, region, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, map[string]any{
		"items": stores,
		"total": total,
		"page":  page,
		"size":  pageSize,
	})
}

// Get 获取门店详情
func (h *StoreHandler) Get(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	store, err := h.svc.GetStore(c.Context(), uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, store)
}

// Create 创建门店
func (h *StoreHandler) Create(c gorp.Context) {
	var req request.CreateStoreRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	s := &biz.Store{
		Name:     req.Name,
		Code:     req.Code,
		Address:  req.Address,
		Phone:    req.Phone,
		Manager:  req.Manager,
		Region:   req.Region,
		Business: req.Business,
		Status:   req.Status,
		Lng:      req.Lng,
		Lat:      req.Lat,
	}
	if err := h.svc.CreateStore(c.Context(), s); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, 201, s)
}

// Update 更新门店
func (h *StoreHandler) Update(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	var req request.UpdateStoreRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	s := &biz.Store{
		ID:       uint(id),
		Name:     req.Name,
		Address:  req.Address,
		Phone:    req.Phone,
		Manager:  req.Manager,
		Region:   req.Region,
		Business: req.Business,
		Status:   req.Status,
		Lng:      req.Lng,
		Lat:      req.Lat,
	}
	if err := h.svc.UpdateStore(c.Context(), s); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, s)
}

// Delete 删除门店
func (h *StoreHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.svc.DeleteStore(c.Context(), uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}