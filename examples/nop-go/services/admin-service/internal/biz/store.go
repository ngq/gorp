// Package biz 门店模块业务层 —— 门店管理的核心领域逻辑
package biz

import "context"

// ==================== 领域实体 ====================

// Store 门店实体 —— 描述系统中的门店信息
type Store struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`         // 门店名称
	Code        string `json:"code"`         // 门店编码，唯一标识
	Address     string `json:"address"`      // 门店地址
	Phone       string `json:"phone"`        // 联系电话
	Manager     string `json:"manager"`      // 店长姓名
	Region      string `json:"region"`       // 所属区域
	Business    string `json:"business"`     // 营业时间
	Status      int    `json:"status"`       // 状态：0=关闭 1=营业
	Lng         float64 `json:"lng"`         // 经度
	Lat         float64 `json:"lat"`         // 纬度
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// StoreRepo 门店仓储接口 —— 定义门店数据访问的契约
type StoreRepo interface {
	// Create 创建门店
	Create(ctx context.Context, s *Store) error
	// GetByID 根据ID获取门店
	GetByID(ctx context.Context, id uint) (*Store, error)
	// GetByCode 根据编码获取门店
	GetByCode(ctx context.Context, code string) (*Store, error)
	// List 获取门店列表，支持状态和区域过滤
	List(ctx context.Context, status int, region string, page, pageSize int) ([]*Store, int64, error)
	// Update 更新门店
	Update(ctx context.Context, s *Store) error
	// Delete 删除门店
	Delete(ctx context.Context, id uint) error
}

// ==================== 用例 ====================

// StoreUseCase 门店模块用例 —— 封装门店管理的业务逻辑
type StoreUseCase struct {
	repo StoreRepo // 门店仓储
}

// NewStoreUseCase 创建门店模块用例
func NewStoreUseCase(repo StoreRepo) *StoreUseCase {
	return &StoreUseCase{repo: repo}
}

// CreateStore 创建门店
func (uc *StoreUseCase) CreateStore(ctx context.Context, s *Store) error {
	return uc.repo.Create(ctx, s)
}

// GetStore 根据ID获取门店
func (uc *StoreUseCase) GetStore(ctx context.Context, id uint) (*Store, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetStoreByCode 根据编码获取门店
func (uc *StoreUseCase) GetStoreByCode(ctx context.Context, code string) (*Store, error) {
	return uc.repo.GetByCode(ctx, code)
}

// ListStores 获取门店列表
func (uc *StoreUseCase) ListStores(ctx context.Context, status int, region string, page, pageSize int) ([]*Store, int64, error) {
	return uc.repo.List(ctx, status, region, page, pageSize)
}

// UpdateStore 更新门店
func (uc *StoreUseCase) UpdateStore(ctx context.Context, s *Store) error {
	return uc.repo.Update(ctx, s)
}

// DeleteStore 删除门店
func (uc *StoreUseCase) DeleteStore(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}
