// Package request 包含交易服务 HTTP 请求结构体定义
// order.go 定义订单/购物车/心愿单/退换货/结账相关请求结构体
package request

// ============================================================================
// 订单相关请求
// ============================================================================

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	UserID        string  `json:"userId" binding:"required"`
	ShippingAddr  string  `json:"shippingAddress"`
	PaymentMethod string  `json:"paymentMethod"`
	TotalAmount   float64 `json:"totalAmount" binding:"required"`
	Currency      string  `json:"currency"`
}

// ListOrdersRequest 订单列表请求
type ListOrdersRequest struct {
	UserID   string `form:"userId" json:"userId"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
}

// ============================================================================
// 购物车相关请求
// ============================================================================

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	UserID    string `json:"userId" binding:"required"`
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

// UpdateCartItemRequest 更新购物车项请求
type UpdateCartItemRequest struct {
	ItemID   string `json:"itemId" binding:"required"`
	Quantity int    `json:"quantity" binding:"required"`
}

// RemoveFromCartRequest 从购物车移除请求
type RemoveFromCartRequest struct {
	ProductID string `json:"productId" binding:"required"`
}

// ============================================================================
// 心愿单相关请求
// ============================================================================

// AddToWishlistRequest 添加到心愿单请求
type AddToWishlistRequest struct {
	UserID      string `json:"userId" binding:"required"`
	ProductID   string `json:"productId" binding:"required"`
	ProductName string `json:"productName"`
}

// RemoveFromWishlistRequest 从心愿单移除请求
type RemoveFromWishlistRequest struct {
	ProductID string `json:"productId" binding:"required"`
}

// ============================================================================
// 结账相关请求
// ============================================================================

// CheckoutRequest 结账请求
type CheckoutRequest struct {
	UserID        string `json:"userId" binding:"required"`
	ShippingAddr  string `json:"shippingAddress" binding:"required"`
	PaymentMethod string `json:"paymentMethod" binding:"required"`
}

// ============================================================================
// 退换货相关请求
// ============================================================================

// CreateReturnRequestRequest 创建退换货请求
type CreateReturnRequestRequest struct {
	OrderID string `json:"orderId" binding:"required"`
	UserID  string `json:"userId" binding:"required"`
	Reason  string `json:"reason" binding:"required"`
}

// ListReturnRequestsRequest 退换货请求列表请求
type ListReturnRequestsRequest struct {
	UserID   string `form:"userId" json:"userId"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
}