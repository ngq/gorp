// Package biz 供应商模块业务层 —— 供应商管理的核心领域逻辑
package biz

import "context"

// ==================== 领域实体 ====================

// Vendor 供应商实体 —— 描述系统中的供应商信息
type Vendor struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`         // 供应商名称
	Code        string `json:"code"`         // 供应商编码，唯一标识
	Contact     string `json:"contact"`      // 联系人
	Phone       string `json:"phone"`        // 联系电话
	Email       string `json:"email"`        // 邮箱
	Address     string `json:"address"`      // 地址
	Category    string `json:"category"`     // 供应商分类
	BankName    string `json:"bank_name"`    // 开户银行
	BankAccount string `json:"bank_account"` // 银行账号
	Status      int    `json:"status"`       // 状态：0=禁用 1=启用
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// VendorRepo 供应商仓储接口 —— 定义供应商数据访问的契约
type VendorRepo interface {
	// Create 创建供应商
	Create(ctx context.Context, v *Vendor) error
	// GetByID 根据ID获取供应商
	GetByID(ctx context.Context, id uint) (*Vendor, error)
	// GetByCode 根据编码获取供应商
	GetByCode(ctx context.Context, code string) (*Vendor, error)
	// List 获取供应商列表，支持分类和状态过滤
	List(ctx context.Context, category string, status int, page, pageSize int) ([]*Vendor, int64, error)
	// Update 更新供应商
	Update(ctx context.Context, v *Vendor) error
	// Delete 删除供应商
	Delete(ctx context.Context, id uint) error
}

// ==================== 用例 ====================

// VendorUseCase 供应商模块用例 —— 封装供应商管理的业务逻辑
type VendorUseCase struct {
	repo VendorRepo // 供应商仓储
}

// NewVendorUseCase 创建供应商模块用例
func NewVendorUseCase(repo VendorRepo) *VendorUseCase {
	return &VendorUseCase{repo: repo}
}

// CreateVendor 创建供应商
func (uc *VendorUseCase) CreateVendor(ctx context.Context, v *Vendor) error {
	return uc.repo.Create(ctx, v)
}

// GetVendor 根据ID获取供应商
func (uc *VendorUseCase) GetVendor(ctx context.Context, id uint) (*Vendor, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetVendorByCode 根据编码获取供应商
func (uc *VendorUseCase) GetVendorByCode(ctx context.Context, code string) (*Vendor, error) {
	return uc.repo.GetByCode(ctx, code)
}

// ListVendors 获取供应商列表
func (uc *VendorUseCase) ListVendors(ctx context.Context, category string, status int, page, pageSize int) ([]*Vendor, int64, error) {
	return uc.repo.List(ctx, category, status, page, pageSize)
}

// UpdateVendor 更新供应商
func (uc *VendorUseCase) UpdateVendor(ctx context.Context, v *Vendor) error {
	return uc.repo.Update(ctx, v)
}

// DeleteVendor 删除供应商
func (uc *VendorUseCase) DeleteVendor(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}
