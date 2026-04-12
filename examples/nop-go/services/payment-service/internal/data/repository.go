// Package data 支付服务数据访问层
package data

import (
	"context"
	"time"

	"nop-go/services/payment-service/internal/models"

	"gorm.io/gorm"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	GetByID(ctx context.Context, id uint64) (*models.Payment, error)
	GetByOrderID(ctx context.Context, orderID uint64) (*models.Payment, error)
	GetByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error)
	Update(ctx context.Context, payment *models.Payment) error
	UpdateStatus(ctx context.Context, id uint64, status string) error
}

type PaymentTransactionRepository interface {
	Create(ctx context.Context, tx *models.PaymentTransaction) error
	GetByPaymentID(ctx context.Context, paymentID uint64) ([]*models.PaymentTransaction, error)
}

type RefundRepository interface {
	Create(ctx context.Context, refund *models.Refund) error
	GetByID(ctx context.Context, id uint64) (*models.Refund, error)
	GetByPaymentID(ctx context.Context, paymentID uint64) ([]*models.Refund, error)
	Update(ctx context.Context, refund *models.Refund) error
}

type paymentRepo struct{ db *gorm.DB }

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, p *models.Payment) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *paymentRepo) GetByID(ctx context.Context, id uint64) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetByOrderID(ctx context.Context, orderID uint64) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) GetByTransactionID(ctx context.Context, transactionID string) (*models.Payment, error) {
	var p models.Payment
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&p).Error
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepo) Update(ctx context.Context, p *models.Payment) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, id uint64, status string) error {
	updates := map[string]interface{}{"status": status}
	if status == "paid" {
		now := time.Now()
		updates["paid_at"] = now
	}
	return r.db.WithContext(ctx).Model(&models.Payment{}).Where("id = ?", id).Updates(updates).Error
}

type paymentTransactionRepo struct{ db *gorm.DB }

func NewPaymentTransactionRepository(db *gorm.DB) PaymentTransactionRepository {
	return &paymentTransactionRepo{db: db}
}

func (r *paymentTransactionRepo) Create(ctx context.Context, t *models.PaymentTransaction) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *paymentTransactionRepo) GetByPaymentID(ctx context.Context, paymentID uint64) ([]*models.PaymentTransaction, error) {
	var list []*models.PaymentTransaction
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&list).Error
	return list, err
}

type refundRepo struct{ db *gorm.DB }

func NewRefundRepository(db *gorm.DB) RefundRepository {
	return &refundRepo{db: db}
}

func (r *refundRepo) Create(ctx context.Context, ref *models.Refund) error {
	return r.db.WithContext(ctx).Create(ref).Error
}

func (r *refundRepo) GetByID(ctx context.Context, id uint64) (*models.Refund, error) {
	var ref models.Refund
	err := r.db.WithContext(ctx).First(&ref, id).Error
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (r *refundRepo) GetByPaymentID(ctx context.Context, paymentID uint64) ([]*models.Refund, error) {
	var list []*models.Refund
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&list).Error
	return list, err
}

func (r *refundRepo) Update(ctx context.Context, ref *models.Refund) error {
	return r.db.WithContext(ctx).Save(ref).Error
}