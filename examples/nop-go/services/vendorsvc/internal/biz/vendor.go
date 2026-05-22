// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Vendor 供应商领域实体。
type Vendor struct {
	ID           uint      // 供应商ID
	Name         string    // 供应商名称
	Email        string    // 供应商邮箱
	Description  string    // 供应商描述
	Active       bool      // 是否启用
	DisplayOrder int       // 显示排序
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// VendorApply 供应商申请领域实体。
type VendorApply struct {
	ID          uint   // 申请ID
	Name        string // 供应商名称
	Email       string // 供应商邮箱
	Description string // 申请描述
	Status      string // 申请状态：pending/approved/rejected
	CreatedAt   time.Time
}

// VendorRepository 供应商仓储接口。
type VendorRepository interface {
	// Create 创建供应商
	Create(ctx context.Context, vendor *Vendor) error
	// GetByID 根据ID获取供应商
	GetByID(ctx context.Context, id uint) (*Vendor, error)
	// List 获取供应商列表
	List(ctx context.Context, page, size int) ([]*Vendor, int64, error)
	// Update 更新供应商
	Update(ctx context.Context, vendor *Vendor) error
	// Delete 删除供应商
	Delete(ctx context.Context, id uint) error
}

// VendorApplyRepository 供应商申请仓储接口。
type VendorApplyRepository interface {
	// Create 创建供应商申请
	Create(ctx context.Context, apply *VendorApply) error
	// GetByID 根据ID获取申请
	GetByID(ctx context.Context, id uint) (*VendorApply, error)
}

// VendorUseCase 供应商用例。
type VendorUseCase struct {
	repo      VendorRepository
	applyRepo VendorApplyRepository
}

// NewVendorUseCase 创建供应商用例。
func NewVendorUseCase(repo VendorRepository, applyRepo VendorApplyRepository) *VendorUseCase {
	return &VendorUseCase{repo: repo, applyRepo: applyRepo}
}

// Create 创建供应商。
func (uc *VendorUseCase) Create(ctx context.Context, name, email, description string, active bool, displayOrder int) (*Vendor, error) {
	vendor := &Vendor{
		Name:         name,
		Email:        email,
		Description:  description,
		Active:       active,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.Create(ctx, vendor); err != nil {
		return nil, err
	}
	return vendor, nil
}

// GetByID 根据ID获取供应商。
func (uc *VendorUseCase) GetByID(ctx context.Context, id uint) (*Vendor, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取供应商列表。
func (uc *VendorUseCase) List(ctx context.Context, page, size int) ([]*Vendor, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新供应商。
func (uc *VendorUseCase) Update(ctx context.Context, id uint, name, email, description string, active bool, displayOrder int) (*Vendor, error) {
	vendor, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	vendor.Name = name
	vendor.Email = email
	vendor.Description = description
	vendor.Active = active
	vendor.DisplayOrder = displayOrder
	vendor.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, vendor); err != nil {
		return nil, err
	}
	return vendor, nil
}

// Delete 删除供应商。
func (uc *VendorUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// GetApply 查询供应商申请状态。
func (uc *VendorUseCase) GetApply(ctx context.Context, id uint) (*VendorApply, error) {
	return uc.applyRepo.GetByID(ctx, id)
}

// SubmitApply 提交供应商申请。
func (uc *VendorUseCase) SubmitApply(ctx context.Context, name, email, description string) (*VendorApply, error) {
	apply := &VendorApply{
		Name:        name,
		Email:       email,
		Description: description,
		Status:      "pending", // 新申请默认为待审核
		CreatedAt:   time.Now(),
	}
	if err := uc.applyRepo.Create(ctx, apply); err != nil {
		return nil, err
	}
	return apply, nil
}