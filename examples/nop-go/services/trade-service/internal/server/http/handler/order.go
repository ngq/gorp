// Package handler 包含交易服务 HTTP 处理器
// order.go 定义订单/购物车/心愿单/退换货/结账相关 handler
package handler

import (
	"net/http"
	"strconv"

	gorp "github.com/ngq/gorp"
	"nop-go/services/trade-service/internal/biz"
	"nop-go/services/trade-service/internal/server/http/request"
	"nop-go/services/trade-service/internal/server/http/response"
	"nop-go/services/trade-service/internal/service"
)

// OrderHandler 订单处理器，通过 Services 容器获取子服务
type OrderHandler struct {
	svc *service.Services
}

// NewOrderHandler 创建订单处理器
func NewOrderHandler(svc *service.Services) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// ============================================================================
// 订单路由 handler
// ============================================================================

// ListOrders 订单列表
// GET /api/v1/orders
func (h *OrderHandler) ListOrders(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	orders, total, err := h.svc.Order.OrderUC.ListOrders(c.Context(), userID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	items := make([]response.OrderResponse, len(orders))
	for i, o := range orders {
		items[i] = toOrderResponse(o)
	}
	gorp.Success(c, response.OrderListResponse{Items: items, Total: total})
}

// GetOrder 订单详情
// GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c gorp.Context) {
	id := c.Param("id")
	order, err := h.svc.Order.OrderUC.GetOrder(c.Context(), id)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toOrderResponse(order))
}

// CreateOrder 创建订单
// POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c gorp.Context) {
	var req request.CreateOrderRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	order := &biz.Order{
		UserID:        req.UserID,
		ShippingAddr:  req.ShippingAddr,
		PaymentMethod: req.PaymentMethod,
		TotalAmount:   req.TotalAmount,
		Currency:      req.Currency,
		Status:        "pending",
	}
	if err := h.svc.Order.OrderUC.CreateOrder(c.Context(), order); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toOrderResponse(order))
}

// DeleteOrder 删除订单
// DELETE /api/v1/orders/:id
func (h *OrderHandler) DeleteOrder(c gorp.Context) {
	id := c.Param("id")
	if err := h.svc.Order.OrderUC.DeleteOrder(c.Context(), id); err != nil {
		gorp.Error(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ============================================================================
// 购物车路由 handler
// ============================================================================

// GetCart 获取购物车
// GET /api/v1/shopping-cart
func (h *OrderHandler) GetCart(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	if userID == "" {
		gorp.BadRequest(c, "缺少用户ID")
		return
	}
	cart, err := h.svc.Order.CartUC.GetCart(c.Context(), userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toCartResponse(cart))
}

// AddToCart 添加到购物车
// POST /api/v1/shopping-cart/add
func (h *OrderHandler) AddToCart(c gorp.Context) {
	var req request.AddToCartRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	item := &biz.CartItem{
		CartID:    "cart-" + req.UserID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
	if err := h.svc.Order.CartUC.AddToCart(c.Context(), item); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, map[string]interface{}{"success": true})
}

// UpdateCartItem 更新购物车项
// POST /api/v1/shopping-cart/update
func (h *OrderHandler) UpdateCartItem(c gorp.Context) {
	var req request.UpdateCartItemRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	item := &biz.CartItem{ID: req.ItemID, Quantity: req.Quantity}
	if err := h.svc.Order.CartUC.UpdateCartItem(c.Context(), item); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, map[string]interface{}{"success": true})
}

// RemoveFromCart 从购物车移除
// DELETE /api/v1/shopping-cart/:productId
func (h *OrderHandler) RemoveFromCart(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	productID := c.Param("productId")
	if err := h.svc.Order.CartUC.RemoveFromCart(c.Context(), "cart-"+userID, productID); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, map[string]interface{}{"success": true})
}

// ============================================================================
// 心愿单路由 handler
// ============================================================================

// GetWishlist 获取心愿单
// GET /api/v1/wishlist
func (h *OrderHandler) GetWishlist(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	if userID == "" {
		gorp.BadRequest(c, "缺少用户ID")
		return
	}
	wishlist, err := h.svc.Order.WishlistUC.GetWishlist(c.Context(), userID)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, toWishlistResponse(wishlist))
}

// AddToWishlist 添加到心愿单
// POST /api/v1/wishlist/add
func (h *OrderHandler) AddToWishlist(c gorp.Context) {
	var req request.AddToWishlistRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	item := &biz.WishlistItem{
		WishlistID:  "wl-" + req.UserID,
		ProductID:   req.ProductID,
		ProductName: req.ProductName,
	}
	if err := h.svc.Order.WishlistUC.AddToWishlist(c.Context(), item); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, map[string]interface{}{"success": true})
}

// RemoveFromWishlist 从心愿单移除
// DELETE /api/v1/wishlist/:productId
func (h *OrderHandler) RemoveFromWishlist(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	productID := c.Param("productId")
	if err := h.svc.Order.WishlistUC.RemoveFromWishlist(c.Context(), "wl-"+userID, productID); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.Success(c, map[string]interface{}{"success": true})
}

// ============================================================================
// 结账路由 handler
// ============================================================================

// Checkout 结账
// POST /api/v1/checkout
func (h *OrderHandler) Checkout(c gorp.Context) {
	var req request.CheckoutRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	order, err := h.svc.Order.CheckoutSvc.Checkout(c.Context(), req.UserID, req.ShippingAddr, req.PaymentMethod)
	if err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toOrderResponse(order))
}

// ============================================================================
// 退换货路由 handler
// ============================================================================

// ListReturnRequests 退换货请求列表
// GET /api/v1/return-requests
func (h *OrderHandler) ListReturnRequests(c gorp.Context) {
	userID := c.DefaultQuery("userId", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	items, total, err := h.svc.Order.ReturnReqUC.ListReturnRequests(c.Context(), userID, page, pageSize)
	if err != nil {
		gorp.Error(c, err)
		return
	}

	respItems := make([]response.ReturnRequestResponse, len(items))
	for i, rr := range items {
		respItems[i] = toReturnRequestResponse(rr)
	}
	gorp.Success(c, response.ReturnRequestListResponse{Items: respItems, Total: total})
}

// CreateReturnRequest 创建退换货请求
// POST /api/v1/return-requests
func (h *OrderHandler) CreateReturnRequest(c gorp.Context) {
	var req request.CreateReturnRequestRequest
	if err := c.BindJSON(&req); err != nil {
		gorp.BadRequest(c, err.Error())
		return
	}
	rr := &biz.ReturnRequest{
		OrderID: req.OrderID,
		UserID:  req.UserID,
		Reason:  req.Reason,
		Status:  "pending",
	}
	if err := h.svc.Order.ReturnReqUC.CreateReturnRequest(c.Context(), rr); err != nil {
		gorp.Error(c, err)
		return
	}
	gorp.SuccessWithStatus(c, http.StatusCreated, toReturnRequestResponse(rr))
}

// ============================================================================
// 响应转换辅助函数
// ============================================================================

func toOrderResponse(o *biz.Order) response.OrderResponse {
	items := make([]response.OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		items[i] = response.OrderItemResponse{
			ID: item.ID, OrderID: item.OrderID, ProductID: item.ProductID,
			ProductName: item.ProductName, Quantity: item.Quantity,
			UnitPrice: item.UnitPrice, Subtotal: item.Subtotal, CreatedAt: item.CreatedAt,
		}
	}
	return response.OrderResponse{
		ID: o.ID, UserID: o.UserID, Status: o.Status,
		TotalAmount: o.TotalAmount, Currency: o.Currency,
		ShippingAddr: o.ShippingAddr, PaymentMethod: o.PaymentMethod,
		Items: items, CreatedAt: o.CreatedAt, UpdatedAt: o.UpdatedAt,
	}
}

func toCartResponse(cart *biz.Cart) response.CartResponse {
	if cart == nil { return response.CartResponse{} }
	items := make([]response.CartItemResponse, len(cart.Items))
	for i, ci := range cart.Items {
		items[i] = response.CartItemResponse{
			ID: ci.ID, CartID: ci.CartID, ProductID: ci.ProductID,
			Quantity: ci.Quantity, AddedAt: ci.AddedAt,
		}
	}
	return response.CartResponse{
		ID: cart.ID, UserID: cart.UserID, Items: items,
		CreatedAt: cart.CreatedAt, UpdatedAt: cart.UpdatedAt,
	}
}

func toWishlistResponse(wl *biz.Wishlist) response.WishlistResponse {
	if wl == nil { return response.WishlistResponse{} }
	items := make([]response.WishlistItemResponse, len(wl.Items))
	for i, wi := range wl.Items {
		items[i] = response.WishlistItemResponse{
			ID: wi.ID, WishlistID: wi.WishlistID, ProductID: wi.ProductID,
			ProductName: wi.ProductName, AddedAt: wi.AddedAt,
		}
	}
	return response.WishlistResponse{
		ID: wl.ID, UserID: wl.UserID, Items: items,
		CreatedAt: wl.CreatedAt, UpdatedAt: wl.UpdatedAt,
	}
}

func toReturnRequestResponse(rr *biz.ReturnRequest) response.ReturnRequestResponse {
	return response.ReturnRequestResponse{
		ID: rr.ID, OrderID: rr.OrderID, UserID: rr.UserID,
		Reason: rr.Reason, Status: rr.Status, RefundAmt: rr.RefundAmt,
		CreatedAt: rr.CreatedAt, UpdatedAt: rr.UpdatedAt,
	}
}
