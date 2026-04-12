// Package pb 支付服务 gRPC 接口定义
package payment

import (
	"context"
)

// PaymentServiceServer 支付服务服务端接口
type PaymentServiceServer interface {
	// CreatePayment 创建支付
	//
	// 中文说明:
	// - SAGA 事务步骤: 创建待支付记录;
	// - 返回支付 ID;
	// - Compensate: CancelPayment。
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)

	// CancelPayment 取消支付
	//
	// 中文说明:
	// - SAGA 补偿操作;
	// - 订单取消时调用。
	CancelPayment(ctx context.Context, req *CancelPaymentRequest) (*CancelPaymentResponse, error)

	// GetPaymentStatus 获取支付状态
	GetPaymentStatus(ctx context.Context, req *GetPaymentStatusRequest) (*GetPaymentStatusResponse, error)

	// ProcessRefund 处理退款
	ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*ProcessRefundResponse, error)
}

// CreatePaymentRequest 创建支付请求
type CreatePaymentRequest struct {
	OrderID        uint64
	CustomerID     uint64
	Amount         float64
	Currency       string
	PaymentMethod  string // Payment.Alipay, Payment.Wechat
	ReturnURL      string
	NotifyURL      string
	TransactionID  string // DTM 事务 ID
}

// CreatePaymentResponse 创建支付响应
type CreatePaymentResponse struct {
	Success        bool
	PaymentID      uint64
	TransactionID  string
	PayURL         string       // 支付链接
	QRCodeURL      string       // 二维码链接
	ErrorMessage   string
}

// CancelPaymentRequest 取消支付请求
type CancelPaymentRequest struct {
	OrderID   uint64
	PaymentID uint64
	Reason    string
}

// CancelPaymentResponse 取消支付响应
type CancelPaymentResponse struct {
	Success      bool
	ErrorMessage string
}

// GetPaymentStatusRequest 获取支付状态请求
type GetPaymentStatusRequest struct {
	OrderID uint64
}

// GetPaymentStatusResponse 获取支付状态响应
type GetPaymentStatusResponse struct {
	PaymentID      uint64
	Status         string // pending, paid, failed, cancelled, refunded
	Amount         float64
	TransactionID  string
	PaidAt         string
}

// ProcessRefundRequest 处理退款请求
type ProcessRefundRequest struct {
	PaymentID uint64
	OrderID   uint64
	Amount    float64
	Reason    string
}

// ProcessRefundResponse 处理退款响应
type ProcessRefundResponse struct {
	Success      bool
	RefundID     string
	ErrorMessage string
}

// UnimplementedPaymentServiceServer 未实现的服务端基类
type UnimplementedPaymentServiceServer struct{}

func (UnimplementedPaymentServiceServer) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) CancelPayment(ctx context.Context, req *CancelPaymentRequest) (*CancelPaymentResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) GetPaymentStatus(ctx context.Context, req *GetPaymentStatusRequest) (*GetPaymentStatusResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*ProcessRefundResponse, error) {
	return nil, nil
}

// PaymentServiceClient 客户端接口
type PaymentServiceClient interface {
	CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error)
	CancelPayment(ctx context.Context, req *CancelPaymentRequest) (*CancelPaymentResponse, error)
	GetPaymentStatus(ctx context.Context, req *GetPaymentStatusRequest) (*GetPaymentStatusResponse, error)
	ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*ProcessRefundResponse, error)
}

// NewPaymentServiceClient 创建客户端
func NewPaymentServiceClient(conn interface{}) PaymentServiceClient {
	return &paymentClient{}
}

type paymentClient struct{}

func (c *paymentClient) CreatePayment(ctx context.Context, req *CreatePaymentRequest) (*CreatePaymentResponse, error) {
	return nil, nil
}
func (c *paymentClient) CancelPayment(ctx context.Context, req *CancelPaymentRequest) (*CancelPaymentResponse, error) {
	return nil, nil
}
func (c *paymentClient) GetPaymentStatus(ctx context.Context, req *GetPaymentStatusRequest) (*GetPaymentStatusResponse, error) {
	return nil, nil
}
func (c *paymentClient) ProcessRefund(ctx context.Context, req *ProcessRefundRequest) (*ProcessRefundResponse, error) {
	return nil, nil
}