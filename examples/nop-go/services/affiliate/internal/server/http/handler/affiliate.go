package handler

import (
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/affiliate/internal/server/http/request"
	"nop-go/services/affiliate/internal/server/http/response"
	"nop-go/services/affiliate/internal/service"
)

// AffiliateHandler 联盟服务 HTTP 处理器。
type AffiliateHandler struct {
	affiliate *service.AffiliateService
}

// NewAffiliateHandler 创建联盟服务处理器。
func NewAffiliateHandler(affiliate *service.AffiliateService) *AffiliateHandler {
	return &AffiliateHandler{affiliate: affiliate}
}

// List 联盟列表。
func (h *AffiliateHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.affiliate.List(c, page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 直接使用 service 层返回的响应类型，无需逐字段转换
	gorp.Success(c, response.AffiliateList{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// Create 创建联盟。
func (h *AffiliateHandler) Create(c gorp.Context) {
	var req request.CreateAffiliate
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, "请求参数无效: "+err.Error())
		return
	}

	aff, err := h.affiliate.Create(c, service.CreateAffiliateRequest{
		Name:   req.Name,
		Url:    req.Url,
		Active: req.Active,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, aff)
}

// Update 更新联盟。
func (h *AffiliateHandler) Update(c gorp.Context) {
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

	aff, err := h.affiliate.Update(c, uint(id), service.UpdateAffiliateRequest{
		Name:   req.Name,
		Url:    req.Url,
		Active: req.Active,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, aff)
}

// Delete 删除联盟。
func (h *AffiliateHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	if err := h.affiliate.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, nil)
}

// ListOrders 联盟关联订单。
func (h *AffiliateHandler) ListOrders(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.affiliate.ListOrders(c, uint(id), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 直接使用 service 层返回的响应类型，无需逐字段转换
	gorp.Success(c, response.AffiliateOrderList{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// ListCustomers 联盟关联客户。
func (h *AffiliateHandler) ListCustomers(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的ID参数")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	items, total, err := h.affiliate.ListCustomers(c, uint(id), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 直接使用 service 层返回的响应类型，无需逐字段转换
	gorp.Success(c, response.AffiliateCustomerList{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}
