// Package grpcsvc 支付服务 gRPC 实现
package grpcsvc

import (
	"context"

	"nop-go/shared/payment"
	"nop-go/services/payment-service/internal/biz"
	"nop-go/services/payment-service/internal/models"
	"nop-go/shared/plugin"
)

// PaymentGRPCServer 支付服务 gRPC 服务端
type PaymentGRPCServer struct {
	payment.UnimplementedPaymentServiceServer
	uc      *biz.PaymentUseCase
	registry *plugin.Registry
}

// NewPaymentGRPCServer 创建支付 gRPC 服务端
func NewPaymentGRPCServer(uc *biz.PaymentUseCase, registry *plugin.Registry) *PaymentGRPCServer {
	return &PaymentGRPCServer{uc: uc, registry: registry}
}

// CreatePayment 创建支付
func (s *PaymentGRPCServer) CreatePayment(ctx context.Context, req *payment.CreatePaymentRequest) (*payment.CreatePaymentResponse, error) {
	result, err := s.uc.ProcessPaymentWithPlugin(ctx, &models.CreatePaymentRequest{
		OrderID:       req.OrderID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		ReturnURL:     req.ReturnURL,
		NotifyURL:     req.NotifyURL,
	})
	if err != nil {
		return &payment.CreatePaymentResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &payment.CreatePaymentResponse{
		Success:       true,
		PaymentID:     result.ID,
		TransactionID: result.TransactionID,
		PayURL:        result.PayURL,
	}, nil
}

// CancelPayment 取消支付
func (s *PaymentGRPCServer) CancelPayment(ctx context.Context, req *payment.CancelPaymentRequest) (*payment.CancelPaymentResponse, error) {
	err := s.uc.MarkAsFailed(ctx, req.PaymentID, "ORDER_CANCELLED", req.Reason)
	if err != nil {
		return &payment.CancelPaymentResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &payment.CancelPaymentResponse{Success: true}, nil
}

// GetPaymentStatus 获取支付状态
func (s *PaymentGRPCServer) GetPaymentStatus(ctx context.Context, req *payment.GetPaymentStatusRequest) (*payment.GetPaymentStatusResponse, error) {
	p, err := s.uc.GetPaymentByOrderID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	paidAt := ""
	if p.PaidAt != nil {
		paidAt = p.PaidAt.Format("2006-01-02 15:04:05")
	}

	return &payment.GetPaymentStatusResponse{
		PaymentID:     p.ID,
		Status:        p.Status,
		Amount:        p.Amount,
		TransactionID: p.TransactionID,
		PaidAt:        paidAt,
	}, nil
}

// ProcessRefund 处理退款
func (s *PaymentGRPCServer) ProcessRefund(ctx context.Context, req *payment.ProcessRefundRequest) (*payment.ProcessRefundResponse, error) {
	result, err := s.uc.RefundWithPlugin(ctx, &models.RefundRequest{
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
		Reason:    req.Reason,
	})
	if err != nil {
		return &payment.ProcessRefundResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &payment.ProcessRefundResponse{
		Success:  true,
		RefundID: result.TransactionID,
	}, nil
}