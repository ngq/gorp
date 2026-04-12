// Package models 订单服务数据模型
package models

import (
	"time"
)

// Order 订单
type Order struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	OrderNumber      string    `gorm:"size:32;uniqueIndex;not null" json:"order_number"`
	CustomerID       uint64    `gorm:"not null;index" json:"customer_id"`
	BillingAddressID uint64    `json:"billing_address_id"`
	ShippingAddressID uint64   `json:"shipping_address_id"`
	Subtotal         float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	DiscountAmount   float64   `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TaxAmount        float64   `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	ShippingAmount   float64   `gorm:"type:decimal(10,2);default:0" json:"shipping_amount"`
	Total            float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	CurrencyCode     string    `gorm:"size:8;default:'CNY'" json:"currency_code"`
	OrderStatus      string    `gorm:"size:16;not null;default:'pending';index" json:"order_status"`
	PaymentStatus    string    `gorm:"size:16;default:'pending'" json:"payment_status"`
	ShippingStatus   string    `gorm:"size:16;default:'not_shipped'" json:"shipping_status"`
	CustomerIP       string    `gorm:"size:64" json:"customer_ip"`
	CustomerNote     string    `gorm:"type:text" json:"customer_note"`
	AdminNote        string    `gorm:"type:text" json:"admin_note"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Items            []OrderItem       `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	BillingAddress   *OrderAddress     `gorm:"foreignKey:OrderID;foreignKey:AddressType" json:"billing_address,omitempty"`
	ShippingAddress  *OrderAddress     `gorm:"foreignKey:OrderID;foreignKey:AddressType" json:"shipping_address,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单商品
type OrderItem struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	OrderID        uint64    `gorm:"not null;index" json:"order_id"`
	ProductID      uint64    `gorm:"not null" json:"product_id"`
	ProductName    string    `gorm:"size:256;not null" json:"product_name"`
	SKU            string    `gorm:"size:64;not null" json:"sku"`
	Quantity       int       `gorm:"not null" json:"quantity"`
	UnitPrice      float64   `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	DiscountAmount float64   `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	TaxAmount      float64   `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	Total          float64   `gorm:"type:decimal(10,2);not null" json:"total"`
	Attributes     string    `gorm:"type:json" json:"attributes"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderItem) TableName() string {
	return "order_items"
}

// OrderAddress 订单地址快照
type OrderAddress struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	OrderID     uint64    `gorm:"not null;index" json:"order_id"`
	AddressType string    `gorm:"size:16;not null" json:"address_type"` // billing, shipping
	FirstName   string    `gorm:"size:64" json:"first_name"`
	LastName    string    `gorm:"size:64" json:"last_name"`
	Email       string    `gorm:"size:128" json:"email"`
	Phone       string    `gorm:"size:32" json:"phone"`
	Company     string    `gorm:"size:128" json:"company"`
	Country     string    `gorm:"size:64" json:"country"`
	State       string    `gorm:"size:64" json:"state"`
	City        string    `gorm:"size:64" json:"city"`
	Address1    string    `gorm:"size:256" json:"address1"`
	Address2    string    `gorm:"size:256" json:"address2"`
	ZipCode     string    `gorm:"size:16" json:"zip_code"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderAddress) TableName() string {
	return "order_addresses"
}

// OrderNote 订单备注
type OrderNote struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	OrderID         uint64    `gorm:"not null;index" json:"order_id"`
	Note            string    `gorm:"type:text;not null" json:"note"`
	IsCustomerVisible bool    `gorm:"default:false" json:"is_customer_visible"`
	CreatedBy       uint64    `json:"created_by"`
	CreatedAt       time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrderNote) TableName() string {
	return "order_notes"
}

// GiftCard 礼品卡
type GiftCard struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	Code           string     `gorm:"size:32;uniqueIndex;not null" json:"code"`
	Amount         float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	RemainingAmount float64   `gorm:"type:decimal(10,2);not null" json:"remaining_amount"`
	CustomerID     uint64     `json:"customer_id"`
	OrderID        uint64     `json:"order_id"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	IsRedeemed     bool       `gorm:"default:false" json:"is_redeemed"`
	PurchasedAt    *time.Time `json:"purchased_at"`
	ActivatedAt    *time.Time `json:"activated_at"`
	RedeemedAt     *time.Time `json:"redeemed_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (GiftCard) TableName() string {
	return "gift_cards"
}

// ReturnRequest 退货请求
type ReturnRequest struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	OrderID         uint64     `gorm:"not null;index" json:"order_id"`
	OrderItemID     uint64     `gorm:"not null" json:"order_item_id"`
	CustomerID      uint64     `gorm:"not null;index" json:"customer_id"`
	Reason          string     `gorm:"size:512" json:"reason"`
	Quantity        int        `gorm:"not null" json:"quantity"`
	RequestedAction string     `gorm:"size:16;default:'refund'" json:"requested_action"` // refund, replace, repair
	Status          string     `gorm:"size:16;default:'pending'" json:"status"` // pending, approved, rejected, processed, cancelled
	RefundAmount    float64    `gorm:"type:decimal(10,2)" json:"refund_amount"`
	ProcessedBy     uint64     `json:"processed_by"`
	ProcessedAt     *time.Time `json:"processed_at"`
	AdminNotes      string     `gorm:"type:text" json:"admin_notes"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (ReturnRequest) TableName() string {
	return "return_requests"
}

// DTO
type CreateOrderRequest struct {
	CustomerID        uint64            `json:"customer_id" binding:"required"`
	BillingAddress    AddressInput      `json:"billing_address" binding:"required"`
	ShippingAddress   AddressInput      `json:"shipping_address" binding:"required"`
	Items             []OrderItemInput  `json:"items" binding:"required,min=1"`
	CouponCode        string            `json:"coupon_code"`
	CustomerNote      string            `json:"customer_note"`
}

type AddressInput struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Phone     string `json:"phone" binding:"required"`
	Company   string `json:"company"`
	Country   string `json:"country" binding:"required"`
	State     string `json:"state" binding:"required"`
	City      string `json:"city" binding:"required"`
	Address1  string `json:"address1" binding:"required"`
	Address2  string `json:"address2"`
	ZipCode   string `json:"zip_code" binding:"required"`
}

type OrderItemInput struct {
	ProductID uint64 `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

type OrderResponse struct {
	ID               uint64            `json:"id"`
	OrderNumber      string            `json:"order_number"`
	CustomerID       uint64            `json:"customer_id"`
	Subtotal         float64           `json:"subtotal"`
	DiscountAmount   float64           `json:"discount_amount"`
	TaxAmount        float64           `json:"tax_amount"`
	ShippingAmount   float64           `json:"shipping_amount"`
	Total            float64           `json:"total"`
	OrderStatus      string            `json:"order_status"`
	PaymentStatus    string            `json:"payment_status"`
	ShippingStatus   string            `json:"shipping_status"`
	Items            []OrderItemResponse `json:"items"`
	CreatedAt        string            `json:"created_at"`
}

type OrderItemResponse struct {
	ID          uint64  `json:"id"`
	ProductID   uint64  `json:"product_id"`
	ProductName string  `json:"product_name"`
	SKU         string  `json:"sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

func ToOrderResponse(o *Order) OrderResponse {
	resp := OrderResponse{
		ID:             o.ID,
		OrderNumber:    o.OrderNumber,
		CustomerID:     o.CustomerID,
		Subtotal:       o.Subtotal,
		DiscountAmount: o.DiscountAmount,
		TaxAmount:      o.TaxAmount,
		ShippingAmount: o.ShippingAmount,
		Total:          o.Total,
		OrderStatus:    o.OrderStatus,
		PaymentStatus:  o.PaymentStatus,
		ShippingStatus: o.ShippingStatus,
		CreatedAt:      o.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if len(o.Items) > 0 {
		resp.Items = make([]OrderItemResponse, len(o.Items))
		for i, item := range o.Items {
			resp.Items[i] = OrderItemResponse{
				ID:          item.ID,
				ProductID:   item.ProductID,
				ProductName: item.ProductName,
				SKU:         item.SKU,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				Total:       item.Total,
			}
		}
	}

	return resp
}