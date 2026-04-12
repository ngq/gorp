// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Order 订单领域实体。
type Order struct {
	ID          uint
	UserID      uint
	ProductID   uint
	ProductName string
	Quantity    int
	TotalPrice  float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// OrderRepository 订单仓储接口。
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id uint) (*Order, error)
	List(ctx context.Context, page, size int) ([]*Order, int64, error)
	ListByUserID(ctx context.Context, userID uint, page, size int) ([]*Order, int64, error)
	Delete(ctx context.Context, id uint) error
}

// OrderUseCase 订单用例。
type OrderUseCase struct {
	repo OrderRepository
}

// NewOrderUseCase 创建订单用例。
func NewOrderUseCase(repo OrderRepository) *OrderUseCase {
	return &OrderUseCase{repo: repo}
}

// Create 创建订单。
func (uc *OrderUseCase) Create(ctx context.Context, userID, productID uint, productName string, quantity int, totalPrice float64) (*Order, error) {
	order := &Order{
		UserID:      userID,
		ProductID:   productID,
		ProductName: productName,
		Quantity:    quantity,
		TotalPrice:  totalPrice,
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// GetByID 根据ID获取订单。
func (uc *OrderUseCase) GetByID(ctx context.Context, id uint) (*Order, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取订单列表。
func (uc *OrderUseCase) List(ctx context.Context, page, size int) ([]*Order, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// ListByUserID 获取用户订单列表。
func (uc *OrderUseCase) ListByUserID(ctx context.Context, userID uint, page, size int) ([]*Order, int64, error) {
	return uc.repo.ListByUserID(ctx, userID, page, size)
}

// Delete 删除订单。
func (uc *OrderUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}