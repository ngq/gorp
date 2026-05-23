// Package response 包含交易服务 HTTP 响应结构体定义
// payment.go 定义支付/支付方式相关响应结构体
package response

import "time"

// PaymentResponse 支付记录响应
type PaymentResponse struct {
	ID              string    `json:"id"`
	OrderID         string    `json:"orderId"`
	UserID          string    `json:"userId"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Method          string    `json:"method"`
	Status          string    `json:"status"`
	TransactionID   string    `json:"transactionId"`
	PaymentMethodID string    `json:"paymentMethodId"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// PaymentListResponse 支付记录列表响应
type PaymentListResponse struct {
	Items []PaymentResponse `json:"items"`
	Total int64             `json:"total"`
}

// PaymentMethodResponse 支付方式响应
type PaymentMethodResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Provider  string    `json:"provider"`
	Last4     string    `json:"last4"`
	IsDefault bool      `json:"isDefault"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
