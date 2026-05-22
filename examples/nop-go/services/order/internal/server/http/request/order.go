// Package request 请求结构体定义。
//
// 本文件包含订单服务所有 HTTP 请求结构体。
// 每个请求结构体对应一个 API 端点的入参。
package request

// ============================================================================
// 订单相关请求
// ============================================================================

// ListOrdersRequest 订单列表请求。
//
// 对应 GET /api/v1/orders
type ListOrdersRequest struct {
	Page     int    `form:"page" json:"page"`           // 页码，默认1
	PageSize int    `form:"page_size" json:"page_size"` // 每页条数，默认10
	UserID   uint   `form:"user_id" json:"user_id"`     // 按用户ID过滤（可选）
	Status   string `form:"status" json:"status"`       // 按状态过滤（可选）
}

// GetOrderRequest 订单详情请求。
//
// 对应 GET /api/v1/orders/:id
type GetOrderRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// CreateOrderRequest 创建订单请求。
//
// 对应 POST /api/v1/orders
type CreateOrderRequest struct {
	UserID         uint    `json:"user_id" binding:"required"`         // 下单用户ID
	SubTotal       float64 `json:"sub_total" binding:"required"`       // 商品小计
	ShippingTotal  float64 `json:"shipping_total"`                     // 运费总额
	TaxTotal       float64 `json:"tax_total"`                          // 税费总额
	OrderTotal     float64 `json:"order_total" binding:"required"`     // 订单总额
	CurrencyCode   string  `json:"currency_code"`                      // 货币代码，默认CNY
	BillingAddrID  *uint   `json:"billing_address_id"`                 // 账单地址ID
	ShippingAddrID *uint   `json:"shipping_address_id"`                // 配送地址ID
	ShippingMethod string  `json:"shipping_method"`                    // 配送方式
	PaymentMethod  string  `json:"payment_method"`                     // 支付方式
	CouponCode     string  `json:"coupon_code"`                        // 折扣码
	GiftCardCode   string  `json:"gift_card_code"`                     // 礼品卡代码
}

// DeleteOrderRequest 删除订单请求。
//
// 对应 DELETE /api/v1/orders/:id
type DeleteOrderRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// CancelOrderRequest 取消订单请求。
//
// 对应 POST /api/v1/orders/:id/cancel
type CancelOrderRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// RefundOrderRequest 退款请求。
//
// 对应 POST /api/v1/orders/:id/refund
type RefundOrderRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// PDFInvoiceRequest PDF发票请求。
//
// 对应 GET /api/v1/orders/:id/pdf
type PDFInvoiceRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// RePostPaymentRequest 重新提交支付请求。
//
// 对应 POST /api/v1/orders/:id/repost-payment
type RePostPaymentRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 订单ID
}

// ShipmentDetailRequest 配送详情请求。
//
// 对应 GET /api/v1/orders/:id/shipment/:shipmentId
type ShipmentDetailRequest struct {
	ID         uint `uri:"id" binding:"required" json:"id"`                   // 订单ID
	ShipmentID uint `uri:"shipmentId" binding:"required" json:"shipment_id"`  // 配送ID
}

// ReorderRequest 重新下单请求。
//
// 对应 POST /api/v1/orders/:id/reorder
type ReorderRequest struct {
	ID uint `uri:"id" binding:"required" json:"id"` // 原订单ID
}

// ============================================================================
// 购物车相关请求
// ============================================================================

// GetShoppingCartRequest 获取购物车请求。
//
// 对应 GET /api/v1/shopping-cart
type GetShoppingCartRequest struct {
	UserID uint `form:"user_id" binding:"required" json:"user_id"` // 用户ID
}

// UpdateCartItemRequest 更新购物车项请求。
//
// 对应 POST /api/v1/shopping-cart/update
type UpdateCartItemRequest struct {
	ItemID   uint `json:"item_id" binding:"required"` // 购物车项ID
	Quantity int  `json:"quantity" binding:"required"` // 新数量
}

// AddToCartRequest 添加到购物车请求。
//
// 对应 POST /api/v1/shopping-cart/add
type AddToCartRequest struct {
	UserID      uint    `json:"user_id" binding:"required"`      // 用户ID
	ProductID   uint    `json:"product_id" binding:"required"`   // 商品ID
	ProductName string  `json:"product_name" binding:"required"` // 商品名称
	SKU         string  `json:"sku"`                             // 商品SKU
	Quantity    int     `json:"quantity" binding:"required"`     // 数量
	UnitPrice   float64 `json:"unit_price" binding:"required"`   // 单价
}

// ApplyCouponRequest 应用折扣码请求。
//
// 对应 POST /api/v1/shopping-cart/apply-coupon
type ApplyCouponRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`     // 用户ID
	CouponCode string `json:"coupon_code" binding:"required"` // 折扣码
}

// ApplyGiftCardRequest 应用礼品卡请求。
//
// 对应 POST /api/v1/shopping-cart/apply-gift-card
type ApplyGiftCardRequest struct {
	UserID       uint   `json:"user_id" binding:"required"`        // 用户ID
	GiftCardCode string `json:"gift_card_code" binding:"required"` // 礼品卡代码
}

// ============================================================================
// 愿望清单相关请求
// ============================================================================

// GetWishlistRequest 获取愿望清单请求。
//
// 对应 GET /api/v1/wishlist
type GetWishlistRequest struct {
	UserID uint `form:"user_id" binding:"required" json:"user_id"` // 用户ID
}

// AddToWishlistRequest 添加到愿望清单请求。
//
// 对应 POST /api/v1/wishlist/add
type AddToWishlistRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`      // 用户ID
	ProductID   uint   `json:"product_id" binding:"required"`   // 商品ID
	ProductName string `json:"product_name" binding:"required"` // 商品名称
}

// ============================================================================
// 结账相关请求
// ============================================================================

// GetCheckoutRequest 获取结账信息请求。
//
// 对应 GET /api/v1/checkout
type GetCheckoutRequest struct {
	UserID uint `form:"user_id" binding:"required" json:"user_id"` // 用户ID
}

// BillingAddressRequest 账单地址请求。
//
// 对应 POST /api/v1/checkout/billing-address
type BillingAddressRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`     // 用户ID
	AddressID  uint   `json:"address_id" binding:"required"`  // 地址ID
	FirstName  string `json:"first_name"`                     // 名
	LastName   string `json:"last_name"`                      // 姓
	Email      string `json:"email"`                          // 邮箱
	Phone      string `json:"phone"`                          // 电话
	Address1   string `json:"address1"`                       // 地址行1
	Address2   string `json:"address2"`                       // 地址行2
	City       string `json:"city"`                           // 城市
	State      string `json:"state"`                          // 省/州
	ZipCode    string `json:"zip_code"`                       // 邮编
	Country    string `json:"country"`                        // 国家
}

// ShippingAddressRequest 配送地址请求。
//
// 对应 POST /api/v1/checkout/shipping-address
type ShippingAddressRequest struct {
	UserID     uint   `json:"user_id" binding:"required"`     // 用户ID
	AddressID  uint   `json:"address_id" binding:"required"`  // 地址ID
	FirstName  string `json:"first_name"`                     // 名
	LastName   string `json:"last_name"`                      // 姓
	Phone      string `json:"phone"`                          // 电话
	Address1   string `json:"address1"`                       // 地址行1
	Address2   string `json:"address2"`                       // 地址行2
	City       string `json:"city"`                           // 城市
	State      string `json:"state"`                          // 省/州
	ZipCode    string `json:"zip_code"`                       // 邮编
	Country    string `json:"country"`                        // 国家
}

// ShippingMethodRequest 配送方式请求。
//
// 对应 POST /api/v1/checkout/shipping-method
type ShippingMethodRequest struct {
	UserID         uint    `json:"user_id" binding:"required"`          // 用户ID
	ShippingMethod string  `json:"shipping_method" binding:"required"`  // 配送方式
	ShippingCost   float64 `json:"shipping_cost"`                       // 配送费用
}

// PaymentMethodRequest 支付方式请求。
//
// 对应 POST /api/v1/checkout/payment-method
type PaymentMethodRequest struct {
	UserID        uint    `json:"user_id" binding:"required"`         // 用户ID
	PaymentMethod string  `json:"payment_method" binding:"required"`  // 支付方式
}

// ConfirmCheckoutRequest 确认订单请求。
//
// 对应 POST /api/v1/checkout/confirm
type ConfirmCheckoutRequest struct {
	UserID uint `json:"user_id" binding:"required"` // 用户ID
}

// ============================================================================
// 退货请求相关请求
// ============================================================================

// ListReturnRequestsRequest 退货请求列表请求。
//
// 对应 GET /api/v1/return-requests
type ListReturnRequestsRequest struct {
	Page     int  `form:"page" json:"page"`           // 页码，默认1
	PageSize int  `form:"page_size" json:"page_size"` // 每页条数，默认10
	UserID   uint `form:"user_id" json:"user_id"`     // 按用户ID过滤（可选）
}

// CreateReturnRequestRequest 提交退货请求。
//
// 对应 POST /api/v1/return-requests
type CreateReturnRequestRequest struct {
	OrderID uint    `json:"order_id" binding:"required"`  // 关联订单ID
	UserID  uint    `json:"user_id" binding:"required"`   // 用户ID
	Reason  string  `json:"reason" binding:"required"`    // 退货原因
}
