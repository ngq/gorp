// Package biz 支付服务业务逻辑层
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"nop-go/services/payment-service/internal/data"
	"nop-go/services/payment-service/internal/models"
	"nop-go/shared/plugin"
	shareErrors "nop-go/shared/errors"
)

type PaymentUseCase struct {
	paymentRepo data.PaymentRepository
	txRepo      data.PaymentTransactionRepository
	refundRepo  data.RefundRepository
	registry    *plugin.Registry
}

func NewPaymentUseCase(
	paymentRepo data.PaymentRepository,
	txRepo data.PaymentTransactionRepository,
	refundRepo data.RefundRepository,
	registry *plugin.Registry,
) *PaymentUseCase {
	return &PaymentUseCase{
		paymentRepo: paymentRepo,
		txRepo:      txRepo,
		refundRepo:  refundRepo,
		registry:    registry,
	}
}

func (uc *PaymentUseCase) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.Payment, error) {
	existing, _ := uc.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if existing != nil {
		return nil, errors.New("payment already exists for this order")
	}

	transactionID := generateTransactionID()

	payment := &models.Payment{
		OrderID:       req.OrderID,
		TransactionID: transactionID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		Status:        "pending",
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	return payment, nil
}

func (uc *PaymentUseCase) GetPayment(ctx context.Context, id uint64) (*models.Payment, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, shareErrors.ErrPaymentNotFound
	}
	return payment, nil
}

func (uc *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID uint64) (*models.Payment, error) {
	return uc.paymentRepo.GetByOrderID(ctx, orderID)
}

func (uc *PaymentUseCase) MarkAsPaid(ctx context.Context, paymentID uint64, transactionID string) error {
	return uc.paymentRepo.UpdateStatus(ctx, paymentID, "paid")
}

func (uc *PaymentUseCase) MarkAsFailed(ctx context.Context, paymentID uint64, errorCode, errorMessage string) error {
	return uc.paymentRepo.UpdateStatus(ctx, paymentID, "failed")
}

func (uc *PaymentUseCase) Refund(ctx context.Context, req *models.RefundRequest) (*models.Refund, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, shareErrors.ErrPaymentNotFound
	}

	if payment.Status != "paid" {
		return nil, errors.New("payment not in paid status")
	}

	if req.Amount > payment.Amount-payment.RefundAmount {
		return nil, errors.New("refund amount exceeds available amount")
	}

	refund := &models.Refund{
		PaymentID: req.PaymentID,
		OrderID:   payment.OrderID,
		Amount:    req.Amount,
		Reason:    req.Reason,
		Status:    "pending",
	}

	if err := uc.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	return refund, nil
}

func (uc *PaymentUseCase) ProcessRefund(ctx context.Context, refundID uint64, success bool) error {
	refund, err := uc.refundRepo.GetByID(ctx, refundID)
	if err != nil {
		return err
	}

	if success {
		now := time.Now()
		refund.Status = "success"
		refund.ProcessedAt = &now

		payment, _ := uc.paymentRepo.GetByID(ctx, refund.PaymentID)
		payment.RefundAmount += refund.Amount

		if payment.RefundAmount >= payment.Amount {
			payment.Status = "refunded"
		} else {
			payment.Status = "partial_refund"
		}
		payment.RefundedAt = &now

		uc.paymentRepo.Update(ctx, payment)
	} else {
		refund.Status = "failed"
	}

	return uc.refundRepo.Update(ctx, refund)
}

func generateTransactionID() string {
	return fmt.Sprintf("TXN%s", time.Now().Format("20060102150405"))
}

// ProcessPaymentWithPlugin 使用支付插件处理支付
//
// 中文说明:
// - 通过插件系统处理支付请求;
// - 根据 paymentMethod 查找对应的支付插件;
// - 调用插件的 ProcessPayment 方法创建支付订单;
// - 保存支付记录到数据库。
func (uc *PaymentUseCase) ProcessPaymentWithPlugin(ctx context.Context, req *models.CreatePaymentRequest) (*models.PaymentResponse, error) {
	// 从注册表获取支付插件
	pm, ok := uc.registry.GetPaymentMethod(req.PaymentMethod)
	if !ok {
		return nil, fmt.Errorf("payment method %s not found or not a valid payment plugin", req.PaymentMethod)
	}

	// 检查订单是否已有支付记录
	existing, _ := uc.paymentRepo.GetByOrderID(ctx, req.OrderID)
	if existing != nil {
		return nil, errors.New("payment already exists for this order")
	}

	// 调用插件处理支付
	result, err := pm.ProcessPayment(ctx, &plugin.ProcessPaymentRequest{
		OrderID:     req.OrderID,
		Amount:      req.Amount,
		Currency:    "CNY",
		CustomerID:  0, // 从上下文获取
		ReturnURL:   req.ReturnURL,
		NotifyURL:   req.NotifyURL,
	})
	if err != nil {
		return nil, fmt.Errorf("process payment failed: %w", err)
	}

	// 创建支付记录
	payment := &models.Payment{
		OrderID:       req.OrderID,
		TransactionID: result.TransactionID,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		Status:        "pending",
	}

	if err := uc.paymentRepo.Create(ctx, payment); err != nil {
		return nil, err
	}

	// 返回支付响应
	return &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		TransactionID: payment.TransactionID,
		PaymentMethod: payment.PaymentMethod,
		Amount:        payment.Amount,
		Status:        payment.Status,
		PayURL:        result.RedirectURL,
		CreatedAt:     payment.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

// ListAvailablePaymentMethods 列出可用的支付方式
//
// 中文说明:
// - 返回所有已注册的支付插件;
// - 供前端展示支付方式选择列表。
func (uc *PaymentUseCase) ListAvailablePaymentMethods() []plugin.PaymentMethod {
	return uc.registry.ListPaymentMethods()
}

// RefundWithPlugin 使用支付插件处理退款
//
// 中文说明:
// - 通过插件系统处理退款请求;
// - 调用支付插件的 Refund 方法;
// - 更新支付记录状态。
func (uc *PaymentUseCase) RefundWithPlugin(ctx context.Context, req *models.RefundRequest) (*models.Refund, error) {
	// 获取支付记录
	payment, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		return nil, shareErrors.ErrPaymentNotFound
	}

	if payment.Status != "paid" {
		return nil, errors.New("payment not in paid status")
	}

	if req.Amount > payment.Amount-payment.RefundAmount {
		return nil, errors.New("refund amount exceeds available amount")
	}

	// 获取支付插件
	pm, ok := uc.registry.GetPaymentMethod(payment.PaymentMethod)
	if !ok {
		return nil, fmt.Errorf("payment method %s not found", payment.PaymentMethod)
	}

	// 调用插件退款
	result, err := pm.Refund(ctx, &plugin.RefundRequest{
		PaymentID:      req.PaymentID,
		TransactionID:  payment.TransactionID,
		Amount:         req.Amount,
		Reason:         req.Reason,
	})
	if err != nil {
		return nil, fmt.Errorf("refund failed: %w", err)
	}

	// 创建退款记录
	refund := &models.Refund{
		PaymentID:     req.PaymentID,
		OrderID:       payment.OrderID,
		Amount:        req.Amount,
		Reason:        req.Reason,
		Status:        "success",
		TransactionID: result.RefundTransactionID,
	}

	now := time.Now()
	refund.ProcessedAt = &now

	if err := uc.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}

	// 更新支付记录
	payment.RefundAmount += req.Amount
	if payment.RefundAmount >= payment.Amount {
		payment.Status = "refunded"
	} else {
		payment.Status = "partial_refund"
	}
	payment.RefundedAt = &now
	uc.paymentRepo.Update(ctx, payment)

	return refund, nil
}