// Package request 包含交易服务 HTTP 请求结构体定义
// payment.go 定义支付/支付方式相关请求结构体
package request

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID         string  `json:"orderId" binding:"required"`
	Amount          float64 `json:"amount" binding:"required"`
	Method          string  `json:"method" binding:"required"`
	PaymentMethodID string  `json:"paymentMethodId"`
	Currency        string  `json:"currency"`
}

// UpdatePaymentRequest 更新支付请求
type UpdatePaymentRequest struct {
	Status        string `json:"status" binding:"required"`
	TransactionID string `json:"transactionId"`
}

// CreatePaymentMethodRequest 创建支付方式请求
type CreatePaymentMethodRequest struct {
	UserID    string `json:"userId" binding:"required"`
	Type      string `json:"type" binding:"required"`
	Provider  string `json:"provider"`
	Last4     string `json:"last4"`
	IsDefault bool   `json:"isDefault"`
}

// ListPaymentsRequest 支付列表请求
type ListPaymentsRequest struct {
	UserID   string `form:"userId" json:"userId"`
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
}