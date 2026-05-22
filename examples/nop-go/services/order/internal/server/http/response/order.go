// Package response 响应结构体定义。
//
// 本文件包含订单服务所有 HTTP 响应结构体。
// 每个响应结构体对应一个 API 端点的出参。
package response

// ============================================================================
// 订单相关响应
// ============================================================================

// Order 订单响应。
//
// 用于返回单个订单的完整信息。
type Order struct {
	ID             uint    `json:"id"`                        // 订单ID
	OrderNo        string  `json:"order_no"`                  // 订单编号
	UserID         uint    `json:"user_id"`                   // 下单用户ID
	Status         string  `json:"status"`                    // 订单状态
	SubTotal       float64 `json:"sub_total"`                 // 商品小计
	DiscountTotal  float64 `json:"discount_total"`            // 折扣总额
	ShippingTotal  float64 `json:"shipping_total"`            // 运费总额
	TaxTotal       float64 `json:"tax_total"`                 // 税费总额
	OrderTotal     float64 `json:"order_total"`               // 订单总额
	CurrencyCode   string  `json:"currency_code"`             // 货币代码
	BillingAddrID  *uint   `json:"billing_address_id"`        // 账单地址ID
	ShippingAddrID *uint   `json:"shipping_address_id"`       // 配送地址ID
	ShippingMethod string  `json:"shipping_method"`           // 配送方式
	PaymentMethod  string  `json:"payment_method"`            // 支付方式
	CouponCode     string  `json:"coupon_code"`               // 折扣码
	GiftCardCode   string  `json:"gift_card_code"`            // 礼品卡代码
	CreatedAt      int64   `json:"created_at"`                // 创建时间（Unix时间戳）
	UpdatedAt      int64   `json:"updated_at"`                // 更新时间（Unix时间戳）
}

// OrderList 订单列表响应。
type OrderList struct {
	Items []Order `json:"items"` // 订单列表
	Total int64   `json:"total"` // 总条数
	Page  int     `json:"page"`  // 当前页码
	Size  int     `json:"size"`  // 每页条数
}

// OrderItem 订单项响应。
type OrderItem struct {
	ID          uint    `json:"id"`           // 订单项ID
	OrderID     uint    `json:"order_id"`     // 所属订单ID
	ProductID   uint    `json:"product_id"`   // 商品ID
	ProductName string  `json:"product_name"` // 商品名称
	SKU         string  `json:"sku"`          // 商品SKU
	Quantity    int     `json:"quantity"`     // 数量
	UnitPrice   float64 `json:"unit_price"`   // 单价
	TotalPrice  float64 `json:"total_price"`  // 小计金额
}

// Shipment 配送响应。
type Shipment struct {
	ID             uint    `json:"id"`              // 配送ID
	OrderID        uint    `json:"order_id"`        // 所属订单ID
	TrackingNumber string  `json:"tracking_number"` // 物流追踪号
	ShippingMethod string  `json:"shipping_method"` // 配送方式
	Status         string  `json:"status"`          // 配送状态
	ShippedAt      *int64  `json:"shipped_at"`      // 发货时间（Unix时间戳）
	DeliveredAt    *int64  `json:"delivered_at"`    // 送达时间（Unix时间戳）
	CreatedAt      int64   `json:"created_at"`      // 创建时间（Unix时间戳）
	UpdatedAt      int64   `json:"updated_at"`      // 更新时间（Unix时间戳）
}

// ============================================================================
// 购物车相关响应
// ============================================================================

// ShoppingCartItem 购物车项响应。
type ShoppingCartItem struct {
	ID          uint    `json:"id"`           // 购物车项ID
	ProductID   uint    `json:"product_id"`   // 商品ID
	ProductName string  `json:"product_name"` // 商品名称
	SKU         string  `json:"sku"`          // 商品SKU
	Quantity    int     `json:"quantity"`     // 数量
	UnitPrice   float64 `json:"unit_price"`   // 单价
	TotalPrice  float64 `json:"total_price"`  // 小计金额
}

// ShoppingCart 购物车响应。
type ShoppingCart struct {
	Items      []ShoppingCartItem `json:"items"`        // 购物车项列表
	SubTotal   float64            `json:"sub_total"`    // 商品小计
	TotalItems int                `json:"total_items"`  // 商品总数量
}

// ============================================================================
// 愿望清单相关响应
// ============================================================================

// WishlistItem 愿望清单项响应。
type WishlistItem struct {
	ID          uint   `json:"id"`           // 愿望清单项ID
	ProductID   uint   `json:"product_id"`   // 商品ID
	ProductName string `json:"product_name"` // 商品名称
	CreatedAt   int64  `json:"created_at"`   // 添加时间（Unix时间戳）
}

// Wishlist 愿望清单响应。
type Wishlist struct {
	Items []WishlistItem `json:"items"` // 愿望清单项列表
	Total int            `json:"total"` // 总数量
}

// ============================================================================
// 结账相关响应
// ============================================================================

// Checkout 结账信息响应。
type Checkout struct {
	UserID         uint    `json:"user_id"`          // 用户ID
	BillingAddrID  *uint   `json:"billing_address_id"`  // 账单地址ID
	ShippingAddrID *uint   `json:"shipping_address_id"` // 配送地址ID
	ShippingMethod string  `json:"shipping_method"`  // 配送方式
	PaymentMethod  string  `json:"payment_method"`   // 支付方式
	SubTotal       float64 `json:"sub_total"`        // 商品小计
	ShippingTotal  float64 `json:"shipping_total"`   // 运费总额
	TaxTotal       float64 `json:"tax_total"`        // 税费总额
	OrderTotal     float64 `json:"order_total"`      // 订单总额
}

// CheckoutStepResult 结账步骤操作结果。
//
// 用于结账流程中各步骤的通用响应。
type CheckoutStepResult struct {
	Success bool   `json:"success"` // 是否成功
	Message string `json:"message"` // 提示信息
}

// ConfirmCheckoutResult 确认订单结果。
type ConfirmCheckoutResult struct {
	Success  bool   `json:"success"`   // 是否成功
	OrderID  uint   `json:"order_id"`  // 创建的订单ID
	OrderNo  string `json:"order_no"`  // 订单编号
	Message  string `json:"message"`   // 提示信息
}

// ============================================================================
// 退货请求相关响应
// ============================================================================

// ReturnRequest 退货请求响应。
type ReturnRequest struct {
	ID          uint    `json:"id"`           // 退货请求ID
	OrderID     uint    `json:"order_id"`     // 关联订单ID
	UserID      uint    `json:"user_id"`      // 用户ID
	Reason      string  `json:"reason"`       // 退货原因
	Status      string  `json:"status"`       // 退货状态
	RefundTotal float64 `json:"refund_total"` // 退款金额
	CreatedAt   int64   `json:"created_at"`   // 创建时间（Unix时间戳）
	UpdatedAt   int64   `json:"updated_at"`   // 更新时间（Unix时间戳）
}

// ReturnRequestList 退货请求列表响应。
type ReturnRequestList struct {
	Items []ReturnRequest `json:"items"` // 退货请求列表
	Total int64           `json:"total"` // 总条数
	Page  int             `json:"page"`  // 当前页码
	Size  int             `json:"size"`  // 每页条数
}
