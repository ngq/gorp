// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Affiliate 联盟领域实体。
type Affiliate struct {
	ID        uint      // 联盟ID
	Name      string    // 联盟名称
	Url       string    // 联盟URL
	Active    bool      // 是否启用
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// AffiliateOrder 联盟关联订单领域实体。
type AffiliateOrder struct {
	ID          uint    // 订单ID
	AffiliateID uint    // 联盟ID
	OrderNo     string  // 订单编号
	CustomerID  uint    // 客户ID
	TotalAmount float64 // 订单总金额
	Status      string  // 订单状态
	CreatedAt   time.Time
}

// AffiliateCustomer 联盟关联客户领域实体。
type AffiliateCustomer struct {
	ID          uint   // 客户ID
	AffiliateID uint   // 联盟ID
	Username    string // 客户用户名
	Email       string // 客户邮箱
	CreatedAt   time.Time
}

// AffiliateRepository 联盟仓储接口。
type AffiliateRepository interface {
	// Create 创建联盟
	Create(ctx context.Context, aff *Affiliate) error
	// GetByID 根据ID获取联盟
	GetByID(ctx context.Context, id uint) (*Affiliate, error)
	// List 获取联盟列表
	List(ctx context.Context, page, size int) ([]*Affiliate, int64, error)
	// Update 更新联盟
	Update(ctx context.Context, aff *Affiliate) error
	// Delete 删除联盟
	Delete(ctx context.Context, id uint) error
}

// AffiliateOrderRepository 联盟订单仓储接口。
type AffiliateOrderRepository interface {
	// ListByAffiliateID 根据联盟ID获取关联订单
	ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*AffiliateOrder, int64, error)
}

// AffiliateCustomerRepository 联盟客户仓储接口。
type AffiliateCustomerRepository interface {
	// ListByAffiliateID 根据联盟ID获取关联客户
	ListByAffiliateID(ctx context.Context, affiliateID uint, page, size int) ([]*AffiliateCustomer, int64, error)
}

// AffiliateUseCase 联盟用例。
type AffiliateUseCase struct {
	repo        AffiliateRepository
	orderRepo   AffiliateOrderRepository
	customerRepo AffiliateCustomerRepository
}

// NewAffiliateUseCase 创建联盟用例。
func NewAffiliateUseCase(repo AffiliateRepository, orderRepo AffiliateOrderRepository, customerRepo AffiliateCustomerRepository) *AffiliateUseCase {
	return &AffiliateUseCase{repo: repo, orderRepo: orderRepo, customerRepo: customerRepo}
}

// Create 创建联盟。
func (uc *AffiliateUseCase) Create(ctx context.Context, name, url string, active bool) (*Affiliate, error) {
	aff := &Affiliate{
		Name:      name,
		Url:       url,
		Active:    active,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, aff); err != nil {
		return nil, err
	}
	return aff, nil
}

// GetByID 根据ID获取联盟。
func (uc *AffiliateUseCase) GetByID(ctx context.Context, id uint) (*Affiliate, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取联盟列表。
func (uc *AffiliateUseCase) List(ctx context.Context, page, size int) ([]*Affiliate, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新联盟。
func (uc *AffiliateUseCase) Update(ctx context.Context, id uint, name, url string, active bool) (*Affiliate, error) {
	aff, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	aff.Name = name
	aff.Url = url
	aff.Active = active
	aff.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, aff); err != nil {
		return nil, err
	}
	return aff, nil
}

// Delete 删除联盟。
func (uc *AffiliateUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// ListOrders 获取联盟关联订单。
func (uc *AffiliateUseCase) ListOrders(ctx context.Context, affiliateID uint, page, size int) ([]*AffiliateOrder, int64, error) {
	return uc.orderRepo.ListByAffiliateID(ctx, affiliateID, page, size)
}

// ListCustomers 获取联盟关联客户。
func (uc *AffiliateUseCase) ListCustomers(ctx context.Context, affiliateID uint, page, size int) ([]*AffiliateCustomer, int64, error) {
	return uc.customerRepo.ListByAffiliateID(ctx, affiliateID, page, size)
}