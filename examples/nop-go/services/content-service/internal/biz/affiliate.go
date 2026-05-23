package biz

import (
	"context"
	"time"
)

// ==================== 实体定义 ====================

// Affiliate 推广合作方实体
type Affiliate struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`       // 合作方名称
	Code      string    `json:"code"`       // 合作方编码，唯一标识
	Contact   string    `json:"contact"`    // 联系方式
	Website   string    `json:"website"`    // 合作方网站
	Commission float64   `json:"commission"` // 佣金比例
	Status    string    `json:"status"`     // active / inactive / suspended
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AffiliateOrder 推广订单，记录通过推广合作方产生的订单
type AffiliateOrder struct {
	ID          uint64    `json:"id"`
	AffiliateID uint64    `json:"affiliate_id"` // 关联合作方 ID
	OrderNo     string    `json:"order_no"`     // 订单编号
	Amount      float64   `json:"amount"`       // 订单金额
	Commission  float64   `json:"commission"`   // 佣金金额
	Status      string    `json:"status"`       // pending / confirmed / paid / cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AffiliateCustomer 推广客户，记录通过推广合作方引流来的客户
type AffiliateCustomer struct {
	ID          uint64    `json:"id"`
	AffiliateID uint64    `json:"affiliate_id"` // 关联合作方 ID
	CustomerID  uint64    `json:"customer_id"`  // 客户 ID
	Source      string    `json:"source"`       // 来源渠道
	FirstVisit  time.Time `json:"first_visit"`  // 首次访问时间
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// AffiliateRepo 推广合作方仓储接口
type AffiliateRepo interface {
	// Create 创建推广合作方
	Create(ctx context.Context, affiliate *Affiliate) error
	// GetByID 根据 ID 获取推广合作方
	GetByID(ctx context.Context, id uint64) (*Affiliate, error)
	// GetByCode 根据编码获取推广合作方
	GetByCode(ctx context.Context, code string) (*Affiliate, error)
	// List 获取推广合作方列表
	List(ctx context.Context, offset, limit int) ([]*Affiliate, error)
	// Update 更新推广合作方
	Update(ctx context.Context, affiliate *Affiliate) error
	// Delete 删除推广合作方
	Delete(ctx context.Context, id uint64) error
}

// AffiliateOrderRepo 推广订单仓储接口
type AffiliateOrderRepo interface {
	// ListByAffiliateID 根据合作方 ID 获取订单列表
	ListByAffiliateID(ctx context.Context, affiliateID uint64, offset, limit int) ([]*AffiliateOrder, error)
}

// AffiliateCustomerRepo 推广客户仓储接口
type AffiliateCustomerRepo interface {
	// ListByAffiliateID 根据合作方 ID 获取客户列表
	ListByAffiliateID(ctx context.Context, affiliateID uint64, offset, limit int) ([]*AffiliateCustomer, error)
}

// ==================== 用例 ====================

// AffiliateUseCase 推广合作业务用例
type AffiliateUseCase struct {
	repo       AffiliateRepo
	orderRepo  AffiliateOrderRepo
	custRepo   AffiliateCustomerRepo
}

// NewAffiliateUseCase 创建推广合作用例
func NewAffiliateUseCase(
	repo AffiliateRepo,
	orderRepo AffiliateOrderRepo,
	custRepo AffiliateCustomerRepo,
) *AffiliateUseCase {
	return &AffiliateUseCase{
		repo:      repo,
		orderRepo: orderRepo,
		custRepo:  custRepo,
	}
}

// CreateAffiliate 创建推广合作方
func (uc *AffiliateUseCase) CreateAffiliate(ctx context.Context, affiliate *Affiliate) error {
	return uc.repo.Create(ctx, affiliate)
}

// GetAffiliate 获取推广合作方详情
func (uc *AffiliateUseCase) GetAffiliate(ctx context.Context, id uint64) (*Affiliate, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetAffiliateByCode 根据编码获取推广合作方
func (uc *AffiliateUseCase) GetAffiliateByCode(ctx context.Context, code string) (*Affiliate, error) {
	return uc.repo.GetByCode(ctx, code)
}

// ListAffiliates 获取推广合作方列表
func (uc *AffiliateUseCase) ListAffiliates(ctx context.Context, offset, limit int) ([]*Affiliate, error) {
	return uc.repo.List(ctx, offset, limit)
}

// UpdateAffiliate 更新推广合作方
func (uc *AffiliateUseCase) UpdateAffiliate(ctx context.Context, affiliate *Affiliate) error {
	return uc.repo.Update(ctx, affiliate)
}

// DeleteAffiliate 删除推广合作方
func (uc *AffiliateUseCase) DeleteAffiliate(ctx context.Context, id uint64) error {
	return uc.repo.Delete(ctx, id)
}

// ListOrders 获取推广合作方的订单列表
func (uc *AffiliateUseCase) ListOrders(ctx context.Context, affiliateID uint64, offset, limit int) ([]*AffiliateOrder, error) {
	return uc.orderRepo.ListByAffiliateID(ctx, affiliateID, offset, limit)
}

// ListCustomers 获取推广合作方的客户列表
func (uc *AffiliateUseCase) ListCustomers(ctx context.Context, affiliateID uint64, offset, limit int) ([]*AffiliateCustomer, error) {
	return uc.custRepo.ListByAffiliateID(ctx, affiliateID, offset, limit)
}
