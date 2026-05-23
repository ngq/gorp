// Package biz 包含交易服务的业务逻辑层
// payment.go 定义支付领域实体、仓库接口与支付用例
package biz

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ============================================================================
// 支付领域实体
// ============================================================================

// Payment 支付实体，表示一次支付记录
type Payment struct {
	ID             string         `json:"id"`
	OrderID        string         `json:"orderId"`
	UserID         string         `json:"userId"`
	Amount         float64        `json:"amount"`
	Currency       string         `json:"currency"`
	Method         string         `json:"method"` // credit_card, paypal, bank_transfer
	Status         string         `json:"status"` // pending, completed, failed, refunded
	TransactionID  string         `json:"transactionId"`
	PaymentMethodID string        `json:"paymentMethodId"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// PaymentMethod 支付方式实体，表示用户绑定的支付工具
type PaymentMethod struct {
	ID           string         `json:"id"`
	UserID       string         `json:"userId"`
	Type         string         `json:"type"` // credit_card, paypal, bank_transfer
	Provider     string         `json:"provider"`
	Last4        string         `json:"last4"` // 卡号后4位或账户标识
	IsDefault    bool           `json:"isDefault"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// ============================================================================
// 仓库接口定义
// ============================================================================

// PaymentRepo 支付数据仓库接口
type PaymentRepo interface {
	// Create 创建支付记录
	Create(ctx context.Context, payment *Payment) error
	// GetByID 根据ID获取支付记录
	GetByID(ctx context.Context, id string) (*Payment, error)
	// Update 更新支付记录
	Update(ctx context.Context, payment *Payment) error
	// ListByOrderID 按订单ID查询支付记录
	ListByOrderID(ctx context.Context, orderID string) ([]*Payment, error)
	// ListByUserID 按用户ID分页查询支付记录
	ListByUserID(ctx context.Context, userID string, page, pageSize int) ([]*Payment, int64, error)
}

// PaymentMethodRepo 支付方式数据仓库接口
type PaymentMethodRepo interface {
	// Create 创建支付方式
	Create(ctx context.Context, method *PaymentMethod) error
	// GetByID 根据ID获取支付方式
	GetByID(ctx context.Context, id string) (*PaymentMethod, error)
	// Update 更新支付方式
	Update(ctx context.Context, method *PaymentMethod) error
	// Delete 删除支付方式
	Delete(ctx context.Context, id string) error
	// ListByUserID 按用户ID查询支付方式列表
	ListByUserID(ctx context.Context, userID string) ([]*PaymentMethod, error)
}

// ============================================================================
// 用例（UseCase）
// ============================================================================

// PaymentUseCase 支付业务用例，封装支付记录与支付方式的管理逻辑
type PaymentUseCase struct {
	paymentRepo      PaymentRepo
	paymentMethodRepo PaymentMethodRepo
}

// NewPaymentUseCase 创建支付用例实例
func NewPaymentUseCase(paymentRepo PaymentRepo, paymentMethodRepo PaymentMethodRepo) *PaymentUseCase {
	return &PaymentUseCase{
		paymentRepo:      paymentRepo,
		paymentMethodRepo: paymentMethodRepo,
	}
}

// CreatePayment 创建支付记录
func (uc *PaymentUseCase) CreatePayment(ctx context.Context, payment *Payment) error {
	return uc.paymentRepo.Create(ctx, payment)
}

// GetPayment 获取支付记录详情
func (uc *PaymentUseCase) GetPayment(ctx context.Context, id string) (*Payment, error) {
	payment, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("获取支付记录失败: %w", err)
	}
	return payment, nil
}

// UpdatePayment 更新支付记录（如更新状态）
func (uc *PaymentUseCase) UpdatePayment(ctx context.Context, payment *Payment) error {
	return uc.paymentRepo.Update(ctx, payment)
}

// ListPaymentsByOrder 按订单ID查询支付记录
func (uc *PaymentUseCase) ListPaymentsByOrder(ctx context.Context, orderID string) ([]*Payment, error) {
	return uc.paymentRepo.ListByOrderID(ctx, orderID)
}

// ListPaymentsByUser 按用户ID分页查询支付记录
func (uc *PaymentUseCase) ListPaymentsByUser(ctx context.Context, userID string, page, pageSize int) ([]*Payment, int64, error) {
	return uc.paymentRepo.ListByUserID(ctx, userID, page, pageSize)
}

// CreatePaymentMethod 创建支付方式
func (uc *PaymentUseCase) CreatePaymentMethod(ctx context.Context, method *PaymentMethod) error {
	return uc.paymentMethodRepo.Create(ctx, method)
}

// GetPaymentMethod 获取支付方式详情
func (uc *PaymentUseCase) GetPaymentMethod(ctx context.Context, id string) (*PaymentMethod, error) {
	return uc.paymentMethodRepo.GetByID(ctx, id)
}

// UpdatePaymentMethod 更新支付方式
func (uc *PaymentUseCase) UpdatePaymentMethod(ctx context.Context, method *PaymentMethod) error {
	return uc.paymentMethodRepo.Update(ctx, method)
}

// DeletePaymentMethod 删除支付方式
func (uc *PaymentUseCase) DeletePaymentMethod(ctx context.Context, id string) error {
	return uc.paymentMethodRepo.Delete(ctx, id)
}

// ListPaymentMethods 按用户ID查询支付方式列表
func (uc *PaymentUseCase) ListPaymentMethods(ctx context.Context, userID string) ([]*PaymentMethod, error) {
	return uc.paymentMethodRepo.ListByUserID(ctx, userID)
}
