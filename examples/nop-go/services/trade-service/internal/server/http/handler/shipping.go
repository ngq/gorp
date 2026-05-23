// Package handler 包含交易服务 HTTP 处理器
// shipping.go 定义物流相关 handler
// 注意：原 Provider handler 已重命名为 ShippingProvider handler
package handler

import (
	"net/http"

	gorp "github.com/ngq/gorp"
	"nop-go/services/trade-service/internal/biz"
	"nop-go/services/trade-service/internal/server/http/request"
	"nop-go/services/trade-service/internal/server/http/response"
	"nop-go/services/trade-service/internal/service"
)

// ShippingHandler 物流处理器，通过 Services 容器获取子服务
type ShippingHandler struct {
	svc *service.Services
}

// NewShippingHandler 创建物流处理器
func NewShippingHandler(svc *service.Services) *ShippingHandler {
	return &ShippingHandler{svc: svc}
}

// --- 物流服务商 ---

// ListShippingProviders 获取物流服务商列表
// GET /api/v1/shipping/providers
func (h *ShippingHandler) ListShippingProviders(c gorp.Context) {
	providers, err := h.svc.Shipping.UC.ListShippingProviders(c.Context())
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.ShippingProviderResponse, len(providers))
	for i, p := range providers {
		items[i] = toShippingProviderResponse(p)
	}
	gorp.Success(c, items)
}

// CreateShippingProvider 创建物流服务商
// POST /api/v1/shipping/providers
func (h *ShippingHandler) CreateShippingProvider(c gorp.Context) {
	var req request.CreateShippingProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	provider := &biz.ShippingProvider{
		Name: req.Name, Code: req.Code,
		Description: req.Description, IsActive: req.IsActive,
	}
	if err := h.svc.Shipping.UC.CreateShippingProvider(c.Context(), provider); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toShippingProviderResponse(provider))
}

// UpdateShippingProvider 更新物流服务商
// PUT /api/v1/shipping/providers/:id
func (h *ShippingHandler) UpdateShippingProvider(c gorp.Context) {
	id := c.Param("id")
	var req request.UpdateShippingProviderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	provider, err := h.svc.Shipping.UC.GetShippingProvider(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	provider.Name = req.Name
	provider.Description = req.Description
	provider.IsActive = req.IsActive
	if err := h.svc.Shipping.UC.UpdateShippingProvider(c.Context(), provider); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toShippingProviderResponse(provider))
}

// DeleteShippingProvider 删除物流服务商
// DELETE /api/v1/shipping/providers/:id
func (h *ShippingHandler) DeleteShippingProvider(c gorp.Context) {
	id := c.Param("id")
	if err := h.svc.Shipping.UC.DeleteShippingProvider(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- 物流订单 ---

// CreateShippingOrder 创建物流订单
// POST /api/v1/shipping/orders
func (h *ShippingHandler) CreateShippingOrder(c gorp.Context) {
	var req request.CreateShippingOrderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	order := &biz.ShippingOrder{
		OrderID:            req.OrderID,
		ShippingProviderID: req.ShippingProviderID,
		ShippingAddress:    req.ShippingAddress,
		Status:             "pending",
	}
	if err := h.svc.Shipping.UC.CreateShippingOrder(c.Context(), order); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toShippingOrderResponse(order))
}

// GetShippingOrder 获取物流订单详情
// GET /api/v1/shipping/orders/:id
func (h *ShippingHandler) GetShippingOrder(c gorp.Context) {
	id := c.Param("id")
	order, err := h.svc.Shipping.UC.GetShippingOrder(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toShippingOrderResponse(order))
}

// UpdateShippingOrder 更新物流订单
// PUT /api/v1/shipping/orders/:id
func (h *ShippingHandler) UpdateShippingOrder(c gorp.Context) {
	id := c.Param("id")
	var req request.UpdateShippingOrderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	order, err := h.svc.Shipping.UC.GetShippingOrder(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	order.Status = req.Status
	order.TrackingNumber = req.TrackingNumber
	if err := h.svc.Shipping.UC.UpdateShippingOrder(c.Context(), order); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toShippingOrderResponse(order))
}

// --- 物流事件 ---

// CreateShippingEvent 创建物流事件
// POST /api/v1/shipping/events
func (h *ShippingHandler) CreateShippingEvent(c gorp.Context) {
	var req request.CreateShippingEventRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	event := &biz.ShippingEvent{
		ShippingOrderID: req.ShippingOrderID,
		Status:          req.Status,
		Location:        req.Location,
		Description:     req.Description,
	}
	if err := h.svc.Shipping.UC.CreateShippingEvent(c.Context(), event); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toShippingEventResponse(event))
}

// ListShippingEvents 获取物流事件列表
// GET /api/v1/shipping/events/:shippingOrderId
func (h *ShippingHandler) ListShippingEvents(c gorp.Context) {
	shippingOrderID := c.Param("shippingOrderId")
	events, err := h.svc.Shipping.UC.ListShippingEvents(c.Context(), shippingOrderID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.ShippingEventResponse, len(events))
	for i, e := range events {
		items[i] = toShippingEventResponse(e)
	}
	gorp.Success(c, items)
}

// ============================================================================
// 响应转换辅助函数
// ============================================================================

func toShippingProviderResponse(p *biz.ShippingProvider) response.ShippingProviderResponse {
	return response.ShippingProviderResponse{
		ID: p.ID, Name: p.Name, Code: p.Code,
		Description: p.Description, IsActive: p.IsActive,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toShippingOrderResponse(o *biz.ShippingOrder) response.ShippingOrderResponse {
	// EstimatedDelivery 在 biz 层为 time.Time，响应层为 *time.Time，需取地址
	estimatedDelivery := &o.EstimatedDelivery
	return response.ShippingOrderResponse{
		ID: o.ID, OrderID: o.OrderID, ShippingProviderID: o.ShippingProviderID,
		TrackingNumber: o.TrackingNumber, Status: o.Status,
		ShippingAddress: o.ShippingAddress,
		EstimatedDelivery: estimatedDelivery, ActualDelivery: o.ActualDelivery,
		CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func toShippingEventResponse(e *biz.ShippingEvent) response.ShippingEventResponse {
	return response.ShippingEventResponse{
		ID: e.ID, ShippingOrderID: e.ShippingOrderID,
		Status: e.Status, Location: e.Location,
		Description: e.Description, EventTime: e.EventTime,
		CreatedAt: e.CreatedAt,
	}
}
