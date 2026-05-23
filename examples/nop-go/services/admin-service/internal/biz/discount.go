// Package biz 优惠模块业务层 —— 优惠活动与使用记录的核心领域逻辑
//
// 重要：本文件为从 discount 独立服务合并而来，已重构移除对 request/response 包的依赖。
// UseCase 方法仅操作领域实体（Discount, DiscountUsage），不涉及 DTO 转换。
// 实体 → 响应 DTO 的转换由 service 层负责。
package biz

import "context"

// ==================== 领域实体 ====================

// Discount 优惠实体 —— 描述一个优惠活动/优惠券
type Discount struct {
	ID           uint    `json:"id"`
	Name         string  `json:"name"`          // 优惠名称
	Code         string  `json:"code"`          // 优惠编码，唯一标识
	Type         int     `json:"type"`          // 优惠类型：1=满减 2=折扣 3=固定金额
	Value        float64 `json:"value"`         // 优惠值（根据类型：满减金额/折扣率/固定金额）
	MinAmount    float64 `json:"min_amount"`    // 最低消费金额
	MaxDiscount  float64 `json:"max_discount"`  // 最大优惠金额（折扣类型时生效）
	StartTime    string  `json:"start_time"`    // 生效开始时间
	EndTime      string  `json:"end_time"`      // 生效结束时间
	TotalQuota   int     `json:"total_quota"`   // 总发放数量
	UsedQuota    int     `json:"used_quota"`    // 已使用数量
	PerUserLimit int     `json:"per_user_limit"`// 每人限领数量
	Status       int     `json:"status"`        // 状态：0=禁用 1=启用
	Description  string  `json:"description"`   // 优惠描述
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// DiscountUsage 优惠使用记录实体 —— 记录用户对优惠的使用情况
type DiscountUsage struct {
	ID          uint   `json:"id"`
	DiscountID  uint   `json:"discount_id"`   // 关联优惠ID
	UserID      uint   `json:"user_id"`       // 使用用户ID
	OrderNo     string `json:"order_no"`      // 关联订单号
	UsedAt      string `json:"used_at"`       // 使用时间
	Status      int    `json:"status"`        // 状态：0=未使用 1=已使用 2=已退款
	CreatedAt   string `json:"created_at"`
}

// ==================== 仓储接口 ====================

// DiscountRepo 优惠仓储接口 —— 定义优惠数据访问的契约
type DiscountRepo interface {
	// Create 创建优惠
	Create(ctx context.Context, d *Discount) error
	// GetByID 根据ID获取优惠
	GetByID(ctx context.Context, id uint) (*Discount, error)
	// GetByCode 根据编码获取优惠
	GetByCode(ctx context.Context, code string) (*Discount, error)
	// List 获取优惠列表，支持类型和状态过滤
	List(ctx context.Context, discountType, status int, page, pageSize int) ([]*Discount, int64, error)
	// Update 更新优惠
	Update(ctx context.Context, d *Discount) error
	// Delete 删除优惠
	Delete(ctx context.Context, id uint) error
}

// DiscountUsageRepo 优惠使用记录仓储接口 —— 定义使用记录数据访问的契约
type DiscountUsageRepo interface {
	// Create 创建使用记录
	Create(ctx context.Context, u *DiscountUsage) error
	// GetByID 根据ID获取使用记录
	GetByID(ctx context.Context, id uint) (*DiscountUsage, error)
	// List 获取使用记录列表，支持优惠ID和用户ID过滤
	List(ctx context.Context, discountID, userID uint, page, pageSize int) ([]*DiscountUsage, int64, error)
	// GetUserUsageCount 获取用户对某优惠的使用次数
	GetUserUsageCount(ctx context.Context, discountID, userID uint) (int64, error)
	// Update 更新使用记录
	Update(ctx context.Context, u *DiscountUsage) error
}

// ==================== 用例 ====================

// DiscountUseCase 优惠模块用例 —— 封装优惠与使用记录的业务逻辑
//
// 注意：所有方法返回领域实体，不做 DTO 转换。
// DTO 转换由 service 层（DiscountService）负责。
type DiscountUseCase struct {
	discountRepo DiscountRepo       // 优惠仓储
	usageRepo    DiscountUsageRepo  // 使用记录仓储
}

// NewDiscountUseCase 创建优惠模块用例
func NewDiscountUseCase(discountRepo DiscountRepo, usageRepo DiscountUsageRepo) *DiscountUseCase {
	return &DiscountUseCase{
		discountRepo: discountRepo,
		usageRepo:    usageRepo,
	}
}

// ---------- 优惠用例方法 ----------

// CreateDiscount 创建优惠
func (uc *DiscountUseCase) CreateDiscount(ctx context.Context, d *Discount) error {
	return uc.discountRepo.Create(ctx, d)
}

// GetDiscount 根据ID获取优惠
func (uc *DiscountUseCase) GetDiscount(ctx context.Context, id uint) (*Discount, error) {
	return uc.discountRepo.GetByID(ctx, id)
}

// GetDiscountByCode 根据编码获取优惠
func (uc *DiscountUseCase) GetDiscountByCode(ctx context.Context, code string) (*Discount, error) {
	return uc.discountRepo.GetByCode(ctx, code)
}

// ListDiscounts 获取优惠列表
func (uc *DiscountUseCase) ListDiscounts(ctx context.Context, discountType, status int, page, pageSize int) ([]*Discount, int64, error) {
	return uc.discountRepo.List(ctx, discountType, status, page, pageSize)
}

// UpdateDiscount 更新优惠
func (uc *DiscountUseCase) UpdateDiscount(ctx context.Context, d *Discount) error {
	return uc.discountRepo.Update(ctx, d)
}

// DeleteDiscount 删除优惠
func (uc *DiscountUseCase) DeleteDiscount(ctx context.Context, id uint) error {
	return uc.discountRepo.Delete(ctx, id)
}

// ---------- 使用记录用例方法 ----------

// CreateDiscountUsage 创建优惠使用记录
func (uc *DiscountUseCase) CreateDiscountUsage(ctx context.Context, u *DiscountUsage) error {
	return uc.usageRepo.Create(ctx, u)
}

// GetDiscountUsage 根据ID获取使用记录
func (uc *DiscountUseCase) GetDiscountUsage(ctx context.Context, id uint) (*DiscountUsage, error) {
	return uc.usageRepo.GetByID(ctx, id)
}

// ListDiscountUsages 获取使用记录列表
func (uc *DiscountUseCase) ListDiscountUsages(ctx context.Context, discountID, userID uint, page, pageSize int) ([]*DiscountUsage, int64, error) {
	return uc.usageRepo.List(ctx, discountID, userID, page, pageSize)
}

// GetUserUsageCount 获取用户对某优惠的使用次数
func (uc *DiscountUseCase) GetUserUsageCount(ctx context.Context, discountID, userID uint) (int64, error) {
	return uc.usageRepo.GetUserUsageCount(ctx, discountID, userID)
}
