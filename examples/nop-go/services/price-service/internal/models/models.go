// Package models 价格服务数据模型
package models

import (
	"time"
)

// TaxRate 税率
type TaxRate struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:64;not null" json:"name"`
	CountryCode  string    `gorm:"size:8;index" json:"country_code"`
	StateCode    string    `gorm:"size:8;index" json:"state_code"`
	ZipCode      string    `gorm:"size:16" json:"zip_code"`
	Rate         float64   `gorm:"type:decimal(5,4);not null" json:"rate"` // 0.0875 = 8.75%
	IsDefault    bool      `gorm:"default:false" json:"is_default"`
	Priority     int       `gorm:"default:0" json:"priority"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (TaxRate) TableName() string {
	return "tax_rates"
}

// Discount 折扣
type Discount struct {
	ID             uint64     `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:128;not null" json:"name"`
	DiscountType   string     `gorm:"size:16;not null" json:"discount_type"` // percentage, fixed
	DiscountAmount float64    `gorm:"type:decimal(10,2);not null" json:"discount_amount"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	RequiresCouponCode bool   `gorm:"default:false" json:"requires_coupon_code"`
	CouponCode     string     `gorm:"size:64;uniqueIndex" json:"coupon_code"`
	MinOrderAmount float64    `gorm:"type:decimal(10,2)" json:"min_order_amount"`
	MaxDiscountAmount float64 `gorm:"type:decimal(10,2)" json:"max_discount_amount"`
	UsageLimit     int        `gorm:"default:0" json:"usage_limit"` // 0 = unlimited
	UsedCount      int        `gorm:"default:0" json:"used_count"`
	IsActive       bool       `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Discount) TableName() string {
	return "discounts"
}

// DiscountUsage 折扣使用记录
type DiscountUsage struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	DiscountID uint64    `gorm:"not null;index" json:"discount_id"`
	OrderID    uint64    `gorm:"not null;index" json:"order_id"`
	CustomerID uint64    `gorm:"not null;index" json:"customer_id"`
	UsedAt     time.Time `gorm:"autoCreateTime" json:"used_at"`
}

func (DiscountUsage) TableName() string {
	return "discount_usages"
}

// ProductPrice 商品价格
type ProductPrice struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	ProductID    uint64    `gorm:"not null;uniqueIndex" json:"product_id"`
	BasePrice    float64   `gorm:"type:decimal(10,2);not null" json:"base_price"`
	SalePrice    float64   `gorm:"type:decimal(10,2)" json:"sale_price"`
	CostPrice    float64   `gorm:"type:decimal(10,2)" json:"cost_price"`
	SpecialPrice float64   `gorm:"type:decimal(10,2)" json:"special_price"`
	SpecialStart *time.Time `json:"special_start"`
	SpecialEnd   *time.Time `json:"special_end"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (ProductPrice) TableName() string {
	return "product_prices"
}

// DTO
type CalculatePriceRequest struct {
	ProductID     uint64   `json:"product_id" binding:"required"`
	Quantity      int      `json:"quantity" binding:"required,min=1"`
	CustomerID    uint64   `json:"customer_id"`
	CustomerRoleID uint64  `json:"customer_role_id"`
	CouponCode    string   `json:"coupon_code"`
	CountryCode   string   `json:"country_code"`
	StateCode     string   `json:"state_code"`
}

type PriceResult struct {
	ProductID     uint64  `json:"product_id"`
	BasePrice     float64 `json:"base_price"`
	FinalPrice    float64 `json:"final_price"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxAmount     float64 `json:"tax_amount"`
	Total         float64 `json:"total"`
}

type ApplyCouponRequest struct {
	CouponCode  string    `json:"coupon_code" binding:"required"`
	Subtotal    float64   `json:"subtotal" binding:"required"`
	Items       []OrderItemPrice `json:"items"`
}

type OrderItemPrice struct {
	ProductID uint64 `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Price     float64 `json:"price"`
}

type CouponResponse struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	DiscountType  string  `json:"discount_type"`
	DiscountAmount float64 `json:"discount_amount"`
	MinOrderAmount float64 `json:"min_order_amount"`
	IsValid       bool    `json:"is_valid"`
	Message       string  `json:"message,omitempty"`
}