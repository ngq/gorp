// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Permission 权限领域实体。
type Permission struct {
	ID           uint      // 权限ID
	Name         string    // 权限名称
	SystemName   string    // 权限系统名称（唯一标识）
	Category     string    // 权限分类
	DisplayOrder int       // 显示排序
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// ACLRecord ACL记录领域实体。
type ACLRecord struct {
	ID             uint   // ACL记录ID
	UserID         uint   // 用户ID
	PermissionID   uint   // 权限ID
	PermissionName string // 权限名称（冗余字段，便于展示）
}

// PermissionRepository 权限仓储接口。
type PermissionRepository interface {
	// Create 创建权限
	Create(ctx context.Context, perm *Permission) error
	// GetByID 根据ID获取权限
	GetByID(ctx context.Context, id uint) (*Permission, error)
	// List 获取权限列表
	List(ctx context.Context, page, size int) ([]*Permission, int64, error)
	// Update 更新权限
	Update(ctx context.Context, perm *Permission) error
	// Delete 删除权限
	Delete(ctx context.Context, id uint) error
}

// ACLRepository ACL记录仓储接口。
type ACLRepository interface {
	// Create 创建ACL记录
	Create(ctx context.Context, record *ACLRecord) error
	// List 获取ACL记录列表
	List(ctx context.Context, page, size int) ([]*ACLRecord, int64, error)
	// Delete 删除ACL记录
	Delete(ctx context.Context, id uint) error
}

// PermissionUseCase 权限用例。
type PermissionUseCase struct {
	repo PermissionRepository
}

// NewPermissionUseCase 创建权限用例。
func NewPermissionUseCase(repo PermissionRepository) *PermissionUseCase {
	return &PermissionUseCase{repo: repo}
}

// Create 创建权限。
func (uc *PermissionUseCase) Create(ctx context.Context, name, systemName, category string, displayOrder int) (*Permission, error) {
	perm := &Permission{
		Name:         name,
		SystemName:   systemName,
		Category:     category,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.Create(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// GetByID 根据ID获取权限。
func (uc *PermissionUseCase) GetByID(ctx context.Context, id uint) (*Permission, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取权限列表。
func (uc *PermissionUseCase) List(ctx context.Context, page, size int) ([]*Permission, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新权限。
func (uc *PermissionUseCase) Update(ctx context.Context, id uint, name, systemName, category string, displayOrder int) (*Permission, error) {
	perm, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	perm.Name = name
	perm.SystemName = systemName
	perm.Category = category
	perm.DisplayOrder = displayOrder
	perm.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

// Delete 删除权限。
func (uc *PermissionUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// ACLUseCase ACL用例。
type ACLUseCase struct {
	aclRepo  ACLRepository
	permRepo PermissionRepository
}

// NewACLUseCase 创建ACL用例。
func NewACLUseCase(aclRepo ACLRepository, permRepo PermissionRepository) *ACLUseCase {
	return &ACLUseCase{aclRepo: aclRepo, permRepo: permRepo}
}

// Create 创建ACL记录，同时查询权限名称用于冗余存储。
func (uc *ACLUseCase) Create(ctx context.Context, userID, permissionID uint) (*ACLRecord, error) {
	// 查询权限名称
	perm, err := uc.permRepo.GetByID(ctx, permissionID)
	permName := ""
	if err == nil && perm != nil {
		permName = perm.Name
	}

	record := &ACLRecord{
		UserID:         userID,
		PermissionID:   permissionID,
		PermissionName: permName,
	}
	if err := uc.aclRepo.Create(ctx, record); err != nil {
		return nil, err
	}
	return record, nil
}

// List 获取ACL记录列表。
func (uc *ACLUseCase) List(ctx context.Context, page, size int) ([]*ACLRecord, int64, error) {
	return uc.aclRepo.List(ctx, page, size)
}

// Delete 删除ACL记录。
func (uc *ACLUseCase) Delete(ctx context.Context, id uint) error {
	return uc.aclRepo.Delete(ctx, id)
}