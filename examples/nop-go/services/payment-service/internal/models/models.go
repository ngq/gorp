// Package models 支付服务数据模型
package models

import (
	"time"
)

// Payment 支付记录
type Payment struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	OrderID         uint64     `gorm:"not null;uniqueIndex" json:"order_id"`
	TransactionID   string     `gorm:"size:128;uniqueIndex" json:"transaction_id"`
	PaymentMethod   string     `gorm:"size:32;not null" json:"payment_method"` // alipay, wechat, bank
	Amount          float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	CurrencyCode    string     `gorm:"size:8;default:'CNY'" json:"currency_code"`
	Status          string     `gorm:"size:16;not null;default:'pending'" json:"status"` // pending, authorized, paid, failed, refunded
	PaidAt          *time.Time `json:"paid_at"`
	RefundedAt      *time.Time `json:"refunded_at"`
	RefundAmount    float64    `gorm:"type:decimal(10,2);default:0" json:"refund_amount"`
	NotifyData      string     `gorm:"type:text" json:"notify_data"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}

// PaymentTransaction 支付流水
type PaymentTransaction struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	PaymentID     uint64    `gorm:"not null;index" json:"payment_id"`
	TransactionID string    `gorm:"size:128" json:"transaction_id"`
	Type          string    `gorm:"size:16;not null" json:"type"` // pay, refund, cancel
	Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Status        string    `gorm:"size:16;not null" json:"status"`
	ErrorCode     string    `gorm:"size:32" json:"error_code"`
	ErrorMessage  string    `gorm:"size:512" json:"error_message"`
	RequestData   string    `gorm:"type:text" json:"request_data"`
	ResponseData  string    `gorm:"type:text" json:"response_data"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

// Refund 退款记录
type Refund struct {
	ID              uint64     `gorm:"primaryKey" json:"id"`
	PaymentID       uint64     `gorm:"not null;index" json:"payment_id"`
	OrderID         uint64     `gorm:"not null;index" json:"order_id"`
	Amount          float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Reason          string     `gorm:"size:256" json:"reason"`
	Status          string     `gorm:"size:16;default:'pending'" json:"status"` // pending, processing, success, failed
	TransactionID   string     `gorm:"size:128" json:"transaction_id"`
	ProcessedAt     *time.Time `json:"processed_at"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (Refund) TableName() string {
	return "refunds"
}

// PaymentMethod 支付方式配置
type PaymentMethod struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:64;not null" json:"name"`
	SystemName  string    `gorm:"size:64;uniqueIndex;not null" json:"system_name"`
	Description string    `gorm:"type:text" json:"description"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	DisplayOrder int      `gorm:"default:0" json:"display_order"`
	Config      string    `gorm:"type:json" json:"config"` // JSON 配置
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (PaymentMethod) TableName() string {
	return "payment_methods"
}

// DTO
type CreatePaymentRequest struct {
	OrderID       uint64 `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	Amount        float64 `json:"amount" binding:"required"`
	ReturnURL     string `json:"return_url"`
	NotifyURL     string `json:"notify_url"`
}

type PaymentResponse struct {
	ID            uint64  `json:"id"`
	OrderID       uint64  `json:"order_id"`
	TransactionID string  `json:"transaction_id"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	PayURL        string  `json:"pay_url,omitempty"` // 支付链接
	CreatedAt     string  `json:"created_at"`
}

type RefundRequest struct {
	PaymentID uint64  `json:"payment_id" binding:"required"`
	Amount    float64 `json:"amount" binding:"required"`
	Reason    string  `json:"reason"`
}

func ToPaymentResponse(p *Payment) PaymentResponse {
	return PaymentResponse{
		ID:            p.ID,
		OrderID:       p.OrderID,
		TransactionID: p.TransactionID,
		PaymentMethod: p.PaymentMethod,
		Amount:        p.Amount,
		Status:        p.Status,
		CreatedAt:     p.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}