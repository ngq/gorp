// Package handler 包含交易服务 HTTP 处理器
// payment.go 定义支付/支付方式相关 handler
package handler

import (
	"net/http"

	gorp "github.com/ngq/gorp"
	"nop-go/services/trade-service/internal/biz"
	"nop-go/services/trade-service/internal/server/http/request"
	"nop-go/services/trade-service/internal/server/http/response"
	"nop-go/services/trade-service/internal/service"
)

// PaymentHandler 支付处理器，通过 Services 容器获取子服务
type PaymentHandler struct {
	svc *service.Services
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(svc *service.Services) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// CreatePayment 创建支付记录
// POST /api/v1/payment
func (h *PaymentHandler) CreatePayment(c gorp.Context) {
	var req request.CreatePaymentRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	payment := &biz.Payment{
		OrderID:         req.OrderID,
		Amount:          req.Amount,
		Method:          req.Method,
		PaymentMethodID: req.PaymentMethodID,
		Currency:        req.Currency,
		Status:          "pending",
	}
	if err := h.svc.Payment.UC.CreatePayment(c.Context(), payment); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toPaymentResponse(payment))
}

// GetPayment 获取支付记录详情
// GET /api/v1/payment/:id
func (h *PaymentHandler) GetPayment(c gorp.Context) {
	id := c.Param("id")
	payment, err := h.svc.Payment.UC.GetPayment(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toPaymentResponse(payment))
}

// ListPaymentsByOrder 按订单查询支付记录
// GET /api/v1/payment/order/:orderId
func (h *PaymentHandler) ListPaymentsByOrder(c gorp.Context) {
	orderID := c.Param("orderId")
	payments, err := h.svc.Payment.UC.ListPaymentsByOrder(c.Context(), orderID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.PaymentResponse, len(payments))
	for i, p := range payments {
		items[i] = toPaymentResponse(p)
	}
	gorp.Success(c, response.PaymentListResponse{Items: items, Total: int64(len(items))})
}

// ListPaymentMethods 获取用户支付方式列表
// GET /api/v1/payment/methods
func (h *PaymentHandler) ListPaymentMethods(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	methods, err := h.svc.Payment.UC.ListPaymentMethods(c.Context(), userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	items := make([]response.PaymentMethodResponse, len(methods))
	for i, m := range methods {
		items[i] = toPaymentMethodResponse(m)
	}
	gorp.Success(c, items)
}

// CreatePaymentMethod 创建支付方式
// POST /api/v1/payment/methods
func (h *PaymentHandler) CreatePaymentMethod(c gorp.Context) {
	var req request.CreatePaymentMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	method := &biz.PaymentMethod{
		UserID:    req.UserID,
		Type:      req.Type,
		Provider:  req.Provider,
		Last4:     req.Last4,
		IsDefault: req.IsDefault,
	}
	if err := h.svc.Payment.UC.CreatePaymentMethod(c.Context(), method); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toPaymentMethodResponse(method))
}

// DeletePaymentMethod 删除支付方式
// DELETE /api/v1/payment/methods/:id
func (h *PaymentHandler) DeletePaymentMethod(c gorp.Context) {
	id := c.Param("id")
	if err := h.svc.Payment.UC.DeletePaymentMethod(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================================================================
// 响应转换辅助函数
// ============================================================================

func toPaymentResponse(p *biz.Payment) response.PaymentResponse {
	return response.PaymentResponse{
		ID: p.ID, OrderID: p.OrderID, UserID: p.UserID,
		Amount: p.Amount, Currency: p.Currency, Method: p.Method,
		Status: p.Status, TransactionID: p.TransactionID,
		PaymentMethodID: p.PaymentMethodID,
		CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func toPaymentMethodResponse(m *biz.PaymentMethod) response.PaymentMethodResponse {
	return response.PaymentMethodResponse{
		ID: m.ID, UserID: m.UserID, Type: m.Type,
		Provider: m.Provider, Last4: m.Last4, IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}
