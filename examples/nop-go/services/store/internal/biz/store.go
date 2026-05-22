// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Store 店铺领域实体。
type Store struct {
	ID           uint      // 店铺ID
	Name         string    // 店铺名称
	Url          string    // 店铺URL
	SslEnabled   bool      // 是否启用SSL
	Hosts        string    // 绑定主机列表（逗号分隔）
	DisplayOrder int       // 显示排序
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// StoreRepository 店铺仓储接口。
type StoreRepository interface {
	// Create 创建店铺
	Create(ctx context.Context, store *Store) error
	// GetByID 根据ID获取店铺
	GetByID(ctx context.Context, id uint) (*Store, error)
	// List 获取店铺列表
	List(ctx context.Context, page, size int) ([]*Store, int64, error)
	// Update 更新店铺
	Update(ctx context.Context, store *Store) error
	// Delete 删除店铺
	Delete(ctx context.Context, id uint) error
}

// StoreUseCase 店铺用例。
type StoreUseCase struct {
	repo StoreRepository
}

// NewStoreUseCase 创建店铺用例。
func NewStoreUseCase(repo StoreRepository) *StoreUseCase {
	return &StoreUseCase{repo: repo}
}

// Create 创建店铺。
func (uc *StoreUseCase) Create(ctx context.Context, name, url string, sslEnabled bool, hosts string, displayOrder int) (*Store, error) {
	store := &Store{
		Name:         name,
		Url:          url,
		SslEnabled:   sslEnabled,
		Hosts:        hosts,
		DisplayOrder: displayOrder,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.Create(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

// GetByID 根据ID获取店铺。
func (uc *StoreUseCase) GetByID(ctx context.Context, id uint) (*Store, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取店铺列表。
func (uc *StoreUseCase) List(ctx context.Context, page, size int) ([]*Store, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新店铺。
func (uc *StoreUseCase) Update(ctx context.Context, id uint, name, url string, sslEnabled bool, hosts string, displayOrder int) (*Store, error) {
	store, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	store.Name = name
	store.Url = url
	store.SslEnabled = sslEnabled
	store.Hosts = hosts
	store.DisplayOrder = displayOrder
	store.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

// Delete 删除店铺。
func (uc *StoreUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}