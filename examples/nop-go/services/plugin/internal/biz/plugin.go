// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Plugin Plugin领域实体。
type Plugin struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PluginRepository Plugin仓储接口。
type PluginRepository interface {
	Create(ctx context.Context, plugin *Plugin) error
	GetByID(ctx context.Context, id uint) (*Plugin, error)
	GetByUsername(ctx context.Context, username string) (*Plugin, error)
	List(ctx context.Context, page, size int) ([]*Plugin, int64, error)
	Delete(ctx context.Context, id uint) error
}

// PluginUseCase Plugin用例。
type PluginUseCase struct {
	repo PluginRepository
}

// NewPluginUseCase 创建Plugin用例。
func NewPluginUseCase(repo PluginRepository) *PluginUseCase {
	return &PluginUseCase{repo: repo}
}

// Create 创建Plugin。
func (uc *PluginUseCase) Create(ctx context.Context, username, email string) (*Plugin, error) {
	plugin := &Plugin{
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, plugin); err != nil {
		return nil, err
	}
	return plugin, nil
}

// GetByID 根据ID获取Plugin。
func (uc *PluginUseCase) GetByID(ctx context.Context, id uint) (*Plugin, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取Plugin列表。
func (uc *PluginUseCase) List(ctx context.Context, page, size int) ([]*Plugin, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除Plugin。
func (uc *PluginUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}