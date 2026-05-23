// Package response 包含交易服务 HTTP 响应结构体定义
// order.go 定义订单/购物车/心愿单/退换货/结账相关响应结构体
package response

import "time"

// ============================================================================
// 订单相关响应
// ============================================================================

// OrderResponse 订单响应
type OrderResponse struct {
	ID            string           `json:"id"`
	UserID        string           `json:"userId"`
	Status        string           `json:"status"`
	TotalAmount   float64          `json:"totalAmount"`
	Currency      string           `json:"currency"`
	ShippingAddr  string           `json:"shippingAddress"`
	PaymentMethod string           `json:"paymentMethod"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     time.Time        `json:"createdAt"`
	UpdatedAt     time.Time        `json:"updatedAt"`
}

// OrderItemResponse 订单行项目响应
type OrderItemResponse struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"orderId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unitPrice"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"createdAt"`
}

// OrderListResponse 订单列表响应
type OrderListResponse struct {
	Items []OrderResponse `json:"items"`
	Total int64           `json:"total"`
}

// ============================================================================
// 购物车相关响应
// ============================================================================

// CartResponse 购物车响应
type CartResponse struct {
	ID        string            `json:"id"`
	UserID    string            `json:"userId"`
	Items     []CartItemResponse `json:"items"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

// CartItemResponse 购物车行项目响应
type CartItemResponse struct {
	ID        string    `json:"id"`
	CartID    string    `json:"cartId"`
	ProductID string    `json:"productId"`
	Quantity  int       `json:"quantity"`
	AddedAt   time.Time `json:"addedAt"`
}

// ============================================================================
// 心愿单相关响应
// ============================================================================

// WishlistResponse 心愿单响应
type WishlistResponse struct {
	ID        string               `json:"id"`
	UserID    string               `json:"userId"`
	Items     []WishlistItemResponse `json:"items"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
}

// WishlistItemResponse 心愿单行项目响应
type WishlistItemResponse struct {
	ID          string    `json:"id"`
	WishlistID  string    `json:"wishlistId"`
	ProductID   string    `json:"productId"`
	ProductName string    `json:"productName"`
	AddedAt     time.Time `json:"addedAt"`
}

// ============================================================================
// 结账相关响应
// ============================================================================

// CheckoutResponse 结账响应（复用 OrderResponse）
type CheckoutResponse = OrderResponse

// ============================================================================
// 退换货相关响应
// ============================================================================

// ReturnRequestResponse 退换货请求响应
type ReturnRequestResponse struct {
	ID          string    `json:"id"`
	OrderID     string    `json:"orderId"`
	UserID      string    `json:"userId"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	RefundAmt   float64   `json:"refundAmount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ReturnRequestListResponse 退换货请求列表响应
type ReturnRequestListResponse struct {
	Items []ReturnRequestResponse `json:"items"`
	Total int64                   `json:"total"`
}
