// Package biz 插件模块业务层 —— 插件管理的核心领域逻辑
package biz

import "context"

// ==================== 领域实体 ====================

// Plugin 插件实体 —— 描述系统中的扩展插件
type Plugin struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`         // 插件名称
	Code        string `json:"code"`         // 插件编码，唯一标识
	Version     string `json:"version"`      // 插件版本
	Description string `json:"description"`  // 插件描述
	Author      string `json:"author"`       // 作者
	Config      string `json:"config"`       // 插件配置（JSON格式）
	Status      int    `json:"status"`       // 状态：0=禁用 1=启用
	Sort        int    `json:"sort"`         // 排序权重
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ==================== 仓储接口 ====================

// PluginRepo 插件仓储接口 —— 定义插件数据访问的契约
type PluginRepo interface {
	// Create 创建插件
	Create(ctx context.Context, p *Plugin) error
	// GetByID 根据ID获取插件
	GetByID(ctx context.Context, id uint) (*Plugin, error)
	// GetByCode 根据编码获取插件
	GetByCode(ctx context.Context, code string) (*Plugin, error)
	// List 获取插件列表，支持状态过滤
	List(ctx context.Context, status int, page, pageSize int) ([]*Plugin, int64, error)
	// Update 更新插件
	Update(ctx context.Context, p *Plugin) error
	// Delete 删除插件
	Delete(ctx context.Context, id uint) error
}

// ==================== 用例 ====================

// PluginUseCase 插件模块用例 —— 封装插件管理的业务逻辑
type PluginUseCase struct {
	repo PluginRepo // 插件仓储
}

// NewPluginUseCase 创建插件模块用例
func NewPluginUseCase(repo PluginRepo) *PluginUseCase {
	return &PluginUseCase{repo: repo}
}

// CreatePlugin 创建插件
func (uc *PluginUseCase) CreatePlugin(ctx context.Context, p *Plugin) error {
	return uc.repo.Create(ctx, p)
}

// GetPlugin 根据ID获取插件
func (uc *PluginUseCase) GetPlugin(ctx context.Context, id uint) (*Plugin, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetPluginByCode 根据编码获取插件
func (uc *PluginUseCase) GetPluginByCode(ctx context.Context, code string) (*Plugin, error) {
	return uc.repo.GetByCode(ctx, code)
}

// ListPlugins 获取插件列表
func (uc *PluginUseCase) ListPlugins(ctx context.Context, status int, page, pageSize int) ([]*Plugin, int64, error) {
	return uc.repo.List(ctx, status, page, pageSize)
}

// UpdatePlugin 更新插件
func (uc *PluginUseCase) UpdatePlugin(ctx context.Context, p *Plugin) error {
	return uc.repo.Update(ctx, p)
}

// DeletePlugin 删除插件
func (uc *PluginUseCase) DeletePlugin(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}
