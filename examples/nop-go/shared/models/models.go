// Package models 提供公共数据模型
//
// 中文说明：
// - 定义跨服务共享的数据模型；
// - 包括地址、金额、分页等通用模型。
package models

import "time"

// BaseEntity 基础实体
type BaseEntity struct {
	ID        uint64    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Address 地址模型
//
// 中文说明：
// - 用于客户地址、账单地址、配送地址；
// - 支持国际地址格式。
type Address struct {
	ID          uint64 `json:"id" gorm:"primaryKey"`
	FirstName   string `json:"first_name" gorm:"size:64"`
	LastName    string `json:"last_name" gorm:"size:64"`
	Email       string `json:"email" gorm:"size:128"`
	Phone       string `json:"phone" gorm:"size:32"`
	Company     string `json:"company" gorm:"size:128"`
	Country     string `json:"country" gorm:"size:64"`
	CountryCode string `json:"country_code" gorm:"size:8"`
	State       string `json:"state" gorm:"size:64"`
	StateCode   string `json:"state_code" gorm:"size:8"`
	City        string `json:"city" gorm:"size:64"`
	Address1    string `json:"address1" gorm:"size:256;not null"`
	Address2    string `json:"address2" gorm:"size:256"`
	ZipCode     string `json:"zip_code" gorm:"size:16"`
}

// FullName 返回完整姓名
func (a *Address) FullName() string {
	if a.FirstName != "" && a.LastName != "" {
		return a.FirstName + " " + a.LastName
	}
	if a.FirstName != "" {
		return a.FirstName
	}
	return a.LastName
}

// FullAddress 返回完整地址
func (a *Address) FullAddress() string {
	parts := []string{a.Address1}
	if a.Address2 != "" {
		parts = append(parts, a.Address2)
	}
	if a.City != "" {
		parts = append(parts, a.City)
	}
	if a.State != "" {
		parts = append(parts, a.State)
	}
	if a.ZipCode != "" {
		parts = append(parts, a.ZipCode)
	}
	if a.Country != "" {
		parts = append(parts, a.Country)
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}

// Money 金额模型
//
// 中文说明：
// - 使用 int64 存储最小单位（分）避免浮点精度问题；
// - 支持多货币。
type Money struct {
	Amount   int64  `json:"amount"`             // 金额（分）
	Currency string `json:"currency"`           // 货币代码（CNY、USD、EUR）
}

// NewMoney 创建金额
func NewMoney(amount float64, currency string) Money {
	return Money{
		Amount:   int64(amount * 100),
		Currency: currency,
	}
}

// NewMoneyFromCents 从分创建金额
func NewMoneyFromCents(cents int64, currency string) Money {
	return Money{
		Amount:   cents,
		Currency: currency,
	}
}

// Float64 返回浮点数金额
func (m Money) Float64() float64 {
	return float64(m.Amount) / 100
}

// Yuan 返回元（CNY）
func (m Money) Yuan() float64 {
	return m.Float64()
}

// Cents 返回分
func (m Money) Cents() int64 {
	return m.Amount
}

// Add 加法
func (m Money) Add(other Money) Money {
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}
}

// Sub 减法
func (m Money) Sub(other Money) Money {
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}
}

// Mul 乘法
func (m Money) Mul(n int64) Money {
	return Money{
		Amount:   m.Amount * n,
		Currency: m.Currency,
	}
}

// IsNegative 是否为负数
func (m Money) IsNegative() bool {
	return m.Amount < 0
}

// IsZero 是否为零
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// PagingRequest 分页请求
type PagingRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"size" json:"size"`
}

// SetDefaults 设置默认值
func (p *PagingRequest) SetDefaults() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算偏移量
func (p *PagingRequest) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// PagingResponse 分页响应
type PagingResponse struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"size"`
	TotalPages int         `json:"total_pages"`
}

// NewPagingResponse 创建分页响应
func NewPagingResponse(items interface{}, total int64, page, pageSize int) *PagingResponse {
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}
	return &PagingResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// Gender 性别
type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderOther   Gender = "other"
	GenderUnknown Gender = "unknown"
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusComplete   OrderStatus = "complete"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

// PaymentStatus 支付状态
type PaymentStatus string

const (
	PaymentStatusPending          PaymentStatus = "pending"
	PaymentStatusAuthorized       PaymentStatus = "authorized"
	PaymentStatusPaid             PaymentStatus = "paid"
	PaymentStatusPartialRefund    PaymentStatus = "partial_refund"
	PaymentStatusRefunded         PaymentStatus = "refunded"
	PaymentStatusVoided           PaymentStatus = "voided"
)

// ShippingStatus 配送状态
type ShippingStatus string

const (
	ShippingStatusNotRequired ShippingStatus = "not_required"
	ShippingStatusNotShipped  ShippingStatus = "not_shipped"
	ShippingStatusPartial     ShippingStatus = "partial"
	ShippingStatusShipped     ShippingStatus = "shipped"
	ShippingStatusDelivered   ShippingStatus = "delivered"
)

// ProductType 商品类型
type ProductType string

const (
	ProductTypeSimple  ProductType = "simple"
	ProductTypeGrouped ProductType = "grouped"
)

// AttributeControlType 属性控件类型
type AttributeControlType string

const (
	AttrControlDropdownList  AttributeControlType = "dropdown"
	AttrControlRadioList     AttributeControlType = "radio"
	AttrControlCheckboxes    AttributeControlType = "checkbox"
	AttrControlTextBox       AttributeControlType = "textbox"
	AttrControlTextArea      AttributeControlType = "textarea"
	AttrControlDatePicker    AttributeControlType = "datepicker"
	AttrControlFileUpload    AttributeControlType = "fileupload"
	AttrControlColorPicker   AttributeControlType = "color"
	AttrControlImagePicker   AttributeControlType = "image"
)