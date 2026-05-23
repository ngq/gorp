package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/content-service/internal/server/http/request"
	"nop-go/services/content-service/internal/server/http/response"
	"nop-go/services/content-service/internal/service"
)

// AffiliateHandler 推广合作服务 HTTP 处理器
type AffiliateHandler struct {
	affiliate *service.AffiliateService
}

// NewAffiliateHandler 创建推广合作服务处理器
func NewAffiliateHandler(affiliate *service.AffiliateService) *AffiliateHandler {
	return &AffiliateHandler{affiliate: affiliate}
}

// ==================== 推广合作方 ====================

// ListAffiliates 推广合作方列表
func (h *AffiliateHandler) ListAffiliates(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.affiliate.ListAffiliates(c.Context(), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.AffiliateList{Items: items, Total: total, Page: page, Size: size})
}

// CreateAffiliate 创建推广合作方
func (h *AffiliateHandler) CreateAffiliate(c gorp.Context) {
	var req request.CreateAffiliate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	aff, err := h.affiliate.CreateAffiliate(c.Context(), service.AffiliateRequest{
		Name: req.Name, Code: req.Code, Contact: req.Contact,
		Website: req.Website, Commission: req.Commission, Status: req.Status,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, aff)
}

// GetAffiliate 获取推广合作方详情
func (h *AffiliateHandler) GetAffiliate(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	aff, err := h.affiliate.GetAffiliate(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, aff)
}

// UpdateAffiliate 更新推广合作方
func (h *AffiliateHandler) UpdateAffiliate(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	var req request.UpdateAffiliate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}
	aff, err := h.affiliate.UpdateAffiliate(c.Context(), id, service.AffiliateRequest{
		Name: req.Name, Code: req.Code, Contact: req.Contact,
		Website: req.Website, Commission: req.Commission, Status: req.Status,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, aff)
}

// DeleteAffiliate 删除推广合作方
func (h *AffiliateHandler) DeleteAffiliate(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	if err := h.affiliate.DeleteAffiliate(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, nil)
}

// ==================== 推广订单 & 客户 ====================

// ListOrders 推广合作方关联订单
func (h *AffiliateHandler) ListOrders(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.affiliate.ListOrders(c.Context(), id, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.AffiliateOrderList{Items: items, Total: total, Page: page, Size: size})
}

// ListCustomers 推广合作方关联客户
func (h *AffiliateHandler) ListCustomers(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	items, total, err := h.affiliate.ListCustomers(c.Context(), id, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, response.AffiliateCustomerList{Items: items, Total: total, Page: page, Size: size})
}