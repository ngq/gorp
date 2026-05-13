// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Product 商品领域实体。
type Product struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductRepository 商品仓储接口。
type ProductRepository interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, id uint) (*Product, error)
	List(ctx context.Context, page, size int) ([]*Product, int64, error)
	Delete(ctx context.Context, id uint) error
}

// ProductUseCase 商品用例。
type ProductUseCase struct {
	repo ProductRepository
}

// NewProductUseCase 创建商品用例。
func NewProductUseCase(repo ProductRepository) *ProductUseCase {
	return &ProductUseCase{repo: repo}
}

// Create 创建商品。
func (uc *ProductUseCase) Create(ctx context.Context, name, description string, price float64, stock int) (*Product, error) {
	product := &Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := uc.repo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// GetByID 根据ID获取商品。
func (uc *ProductUseCase) GetByID(ctx context.Context, id uint) (*Product, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取商品列表。
func (uc *ProductUseCase) List(ctx context.Context, page, size int) ([]*Product, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除商品。
func (uc *ProductUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}