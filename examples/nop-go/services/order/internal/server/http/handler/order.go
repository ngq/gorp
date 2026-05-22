// Package handler HTTP 处理器。
//
// 本文件包含订单服务所有 HTTP handler 方法。
// 每个 handler 对应一个路由端点，负责参数解析、调用 service 层、构造响应。
// handler 层不再手动构造 response DTO，而是直接透传 service 层返回的 DTO。
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/order/internal/biz"
	"nop-go/services/order/internal/server/http/request"
	"nop-go/services/order/internal/service"
)

// OrderHandler 订单处理器。
//
// 聚合订单、购物车、愿望清单、结账、退货请求所有 handler 方法。
type OrderHandler struct {
	orderSvc *service.OrderService
	cartSvc  *service.CartService
	wlSvc    *service.WishlistService
	coSvc    *service.CheckoutService
	rrSvc    *service.ReturnRequestService
}

// NewOrderHandler 创建订单处理器。
func NewOrderHandler(
	orderSvc *service.OrderService,
	cartSvc *service.CartService,
	wlSvc *service.WishlistService,
	coSvc *service.CheckoutService,
	rrSvc *service.ReturnRequestService,
) *OrderHandler {
	return &OrderHandler{
		orderSvc: orderSvc,
		cartSvc:  cartSvc,
		wlSvc:    wlSvc,
		coSvc:    coSvc,
		rrSvc:    rrSvc,
	}
}

// ============================================================================
// 订单路由 handler
// ============================================================================

// List 订单列表。
//
// GET /api/v1/orders
// 支持分页和按用户ID、状态过滤。
func (h *OrderHandler) List(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	items, total, err := h.orderSvc.List(c, uint(userID), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.OrderListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// GetByID 订单详情。
//
// GET /api/v1/orders/:id
func (h *OrderHandler) GetByID(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	order, err := h.orderSvc.GetByID(c, uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, order)
}

// Create 创建订单。
//
// POST /api/v1/orders
func (h *OrderHandler) Create(c gorp.Context) {
	var req request.CreateOrderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	order, err := h.orderSvc.Create(c, &biz.Order{
		UserID:         req.UserID,
		SubTotal:       req.SubTotal,
		ShippingTotal:  req.ShippingTotal,
		TaxTotal:       req.TaxTotal,
		OrderTotal:     req.OrderTotal,
		CurrencyCode:   req.CurrencyCode,
		BillingAddrID:  req.BillingAddrID,
		ShippingAddrID: req.ShippingAddrID,
		ShippingMethod: req.ShippingMethod,
		PaymentMethod:  req.PaymentMethod,
		CouponCode:     req.CouponCode,
		GiftCardCode:   req.GiftCardCode,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, order)
}

// Delete 删除订单。
//
// DELETE /api/v1/orders/:id
func (h *OrderHandler) Delete(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	if err := h.orderSvc.Delete(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// CancelOrder 取消订单。
//
// POST /api/v1/orders/:id/cancel
func (h *OrderHandler) CancelOrder(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	if err := h.orderSvc.CancelOrder(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "订单已取消",
	})
}

// RefundOrder 退款。
//
// POST /api/v1/orders/:id/refund
func (h *OrderHandler) RefundOrder(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	if err := h.orderSvc.RefundOrder(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "退款已处理",
	})
}

// GetPDFInvoice PDF发票。
//
// GET /api/v1/orders/:id/pdf
func (h *OrderHandler) GetPDFInvoice(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	pdfData, err := h.orderSvc.GetPDFInvoice(c, uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	// 返回 PDF 文件流
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// RePostPayment 重新提交支付。
//
// POST /api/v1/orders/:id/repost-payment
func (h *OrderHandler) RePostPayment(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	if err := h.orderSvc.RePostPayment(c, uint(id)); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "支付已重新提交",
	})
}

// GetShipmentDetail 配送详情。
//
// GET /api/v1/orders/:id/shipment/:shipmentId
func (h *OrderHandler) GetShipmentDetail(c gorp.Context) {
	orderID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}
	shipmentID, err := strconv.ParseUint(c.Param("shipmentId"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的配送ID")
		return
	}

	shipment, err := h.orderSvc.GetShipment(c, uint(orderID), uint(shipmentID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, shipment)
}

// Reorder 重新下单。
//
// POST /api/v1/orders/:id/reorder
func (h *OrderHandler) Reorder(c gorp.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		gorp.BadRequest(c, "无效的订单ID")
		return
	}

	order, err := h.orderSvc.Reorder(c, uint(id))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, order)
}

// ============================================================================
// 购物车路由 handler
// ============================================================================

// GetShoppingCart 获取购物车。
//
// GET /api/v1/shopping-cart
func (h *OrderHandler) GetShoppingCart(c gorp.Context) {
	userID, err := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	if userID == 0 || err != nil {
		gorp.BadRequest(c, "缺少用户ID")
		return
	}

	cart, err := h.cartSvc.GetCart(c, uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, cart)
}

// UpdateCart 更新购物车。
//
// POST /api/v1/shopping-cart/update
func (h *OrderHandler) UpdateCart(c gorp.Context) {
	var req request.UpdateCartItemRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.cartSvc.UpdateCart(c, req.ItemID, req.Quantity); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "购物车已更新",
	})
}

// AddToCart 添加到购物车。
//
// POST /api/v1/shopping-cart/add
func (h *OrderHandler) AddToCart(c gorp.Context) {
	var req request.AddToCartRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.cartSvc.AddToCart(c, &biz.ShoppingCartItem{
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
		SKU:         req.SKU,
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
	}); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, service.StepResult{
		Success: true,
		Message: "已添加到购物车",
	})
}

// ApplyCoupon 应用折扣码。
//
// POST /api/v1/shopping-cart/apply-coupon
func (h *OrderHandler) ApplyCoupon(c gorp.Context) {
	var req request.ApplyCouponRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.cartSvc.ApplyCoupon(c, req.UserID, req.CouponCode); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "折扣码已应用",
	})
}

// ApplyGiftCard 应用礼品卡。
//
// POST /api/v1/shopping-cart/apply-gift-card
func (h *OrderHandler) ApplyGiftCard(c gorp.Context) {
	var req request.ApplyGiftCardRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.cartSvc.ApplyGiftCard(c, req.UserID, req.GiftCardCode); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "礼品卡已应用",
	})
}

// ============================================================================
// 愿望清单路由 handler
// ============================================================================

// GetWishlist 获取愿望清单。
//
// GET /api/v1/wishlist
func (h *OrderHandler) GetWishlist(c gorp.Context) {
	userID, err := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	if userID == 0 || err != nil {
		gorp.BadRequest(c, "缺少用户ID")
		return
	}

	wishlist, err := h.wlSvc.GetWishlist(c, uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, wishlist)
}

// AddToWishlist 添加到愿望清单。
//
// POST /api/v1/wishlist/add
func (h *OrderHandler) AddToWishlist(c gorp.Context) {
	var req request.AddToWishlistRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.wlSvc.AddToWishlist(c, &biz.WishlistItem{
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
	}); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, service.StepResult{
		Success: true,
		Message: "已添加到愿望清单",
	})
}

// ============================================================================
// 结账路由 handler
// ============================================================================

// GetCheckout 获取结账信息。
//
// GET /api/v1/checkout
func (h *OrderHandler) GetCheckout(c gorp.Context) {
	userID, err := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)
	if userID == 0 || err != nil {
		gorp.BadRequest(c, "缺少用户ID")
		return
	}

	checkout, err := h.coSvc.GetCheckout(c, uint(userID))
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, checkout)
}

// SetBillingAddress 设置账单地址。
//
// POST /api/v1/checkout/billing-address
func (h *OrderHandler) SetBillingAddress(c gorp.Context) {
	var req request.BillingAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.coSvc.SetBillingAddress(c, req.UserID, req.AddressID); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "账单地址已设置",
	})
}

// SetShippingAddress 设置配送地址。
//
// POST /api/v1/checkout/shipping-address
func (h *OrderHandler) SetShippingAddress(c gorp.Context) {
	var req request.ShippingAddressRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.coSvc.SetShippingAddress(c, req.UserID, req.AddressID); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "配送地址已设置",
	})
}

// SetShippingMethod 设置配送方式。
//
// POST /api/v1/checkout/shipping-method
func (h *OrderHandler) SetShippingMethod(c gorp.Context) {
	var req request.ShippingMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.coSvc.SetShippingMethod(c, req.UserID, req.ShippingMethod, req.ShippingCost); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "配送方式已设置",
	})
}

// SetPaymentMethod 设置支付方式。
//
// POST /api/v1/checkout/payment-method
func (h *OrderHandler) SetPaymentMethod(c gorp.Context) {
	var req request.PaymentMethodRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	if err := h.coSvc.SetPaymentMethod(c, req.UserID, req.PaymentMethod); err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.StepResult{
		Success: true,
		Message: "支付方式已设置",
	})
}

// ConfirmCheckout 确认订单。
//
// POST /api/v1/checkout/confirm
func (h *OrderHandler) ConfirmCheckout(c gorp.Context) {
	var req request.ConfirmCheckoutRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	result, err := h.coSvc.Confirm(c, req.UserID)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, result)
}

// ============================================================================
// 退货请求路由 handler
// ============================================================================

// ListReturnRequests 退货请求列表。
//
// GET /api/v1/return-requests
func (h *OrderHandler) ListReturnRequests(c gorp.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	userID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	items, total, err := h.rrSvc.List(c, uint(userID), page, size)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.Success(c, service.ReturnRequestListResponse{
		Items: items,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

// CreateReturnRequest 提交退货请求。
//
// POST /api/v1/return-requests
func (h *OrderHandler) CreateReturnRequest(c gorp.Context) {
	var req request.CreateReturnRequestRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}

	rr, err := h.rrSvc.Create(c, &biz.ReturnRequest{
		OrderID: req.OrderID,
		UserID:  req.UserID,
		Reason:  req.Reason,
	})
	if err != nil {
		gorp.Error(c, err)
		return
	}

	gorp.SuccessWithStatus(c, http.StatusCreated, rr)
}