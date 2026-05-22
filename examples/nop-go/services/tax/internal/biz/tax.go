// Package biz 提供 tax 服务的业务逻辑层
//
// 税务服务包含两个子领域：
// 1. 税务提供者（Provider）— 税务计算引擎/服务
// 2. 税类别（Category）— 商品税分类
package biz

import (
	"context"
	"time"

	"nop-go/services/tax/internal/server/http/request"
	"nop-go/services/tax/internal/server/http/response"
)

// ==================== 领域实体定义 ====================

// Provider 税务提供者领域实体
type Provider struct {
	ID            uint      // 提供者 ID
	Name          string    // 提供者名称
	SystemKeyword string    // 系统关键字标识
	DisplayOrder  int       // 显示排序
	IsActive      bool      // 是否启用
	IsPrimary     bool      // 是否为主要税务提供者
	LogoURL       string    // Logo 地址
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// Category 税类别领域实体
type Category struct {
	ID           uint      // 税类别 ID
	Name         string    // 税类别名称
	Rate         float64   // 税率百分比
	DisplayOrder int       // 显示排序
	IsActive     bool      // 是否启用
	Description  string    // 税类别描述
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// ==================== 仓储接口定义 ====================

// ProviderRepository 税务提供者仓储接口
type ProviderRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Provider, int64, error)
	Update(ctx context.Context, provider *Provider) (*Provider, error)
}

// CategoryRepository 税类别仓储接口
type CategoryRepository interface {
	List(ctx context.Context, page, pageSize int) ([]*Category, int64, error)
	Create(ctx context.Context, category *Category) (*Category, error)
	Update(ctx context.Context, category *Category) (*Category, error)
	Delete(ctx context.Context, id uint) error
}

// ==================== 用例实现 ====================

// TaxUsecase 税务服务业务用例
type TaxUsecase struct {
	providerRepo  ProviderRepository
	categoryRepo  CategoryRepository
}

// NewTaxUsecase 创建税务服务业务用例
func NewTaxUsecase(providerRepo ProviderRepository, categoryRepo CategoryRepository) *TaxUsecase {
	return &TaxUsecase{
		providerRepo: providerRepo,
		categoryRepo: categoryRepo,
	}
}

// ==================== 税务提供者用例 ====================

// ListProviders 获取税务提供者列表
func (uc *TaxUsecase) ListProviders(ctx context.Context, page, pageSize int) ([]*response.ProviderResponse, int64, error) {
	providers, total, err := uc.providerRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.ProviderResponse, len(providers))
	for i, p := range providers {
		items[i] = toProviderResponse(p)
	}
	return items, total, nil
}

// UpdateProvider 更新税务提供者
func (uc *TaxUsecase) UpdateProvider(ctx context.Context, req request.UpdateProviderRequest) (*response.ProviderResponse, error) {
	provider := &Provider{
		ID:            req.ID,
		Name:          req.Name,
		SystemKeyword: req.SystemKeyword,
		DisplayOrder:  req.DisplayOrder,
		IsActive:      req.IsActive,
		IsPrimary:     req.IsPrimary,
		LogoURL:       req.LogoURL,
	}

	updated, err := uc.providerRepo.Update(ctx, provider)
	if err != nil {
		return nil, err
	}
	return toProviderResponse(updated), nil
}

// ==================== 税类别用例 ====================

// ListCategories 获取税类别列表
func (uc *TaxUsecase) ListCategories(ctx context.Context, page, pageSize int) ([]*response.CategoryResponse, int64, error) {
	categories, total, err := uc.categoryRepo.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*response.CategoryResponse, len(categories))
	for i, c := range categories {
		items[i] = toCategoryResponse(c)
	}
	return items, total, nil
}

// CreateCategory 创建税类别
func (uc *TaxUsecase) CreateCategory(ctx context.Context, req request.CreateCategoryRequest) (*response.CategoryResponse, error) {
	category := &Category{
		Name:         req.Name,
		Rate:         req.Rate,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
		Description:  req.Description,
	}

	created, err := uc.categoryRepo.Create(ctx, category)
	if err != nil {
		return nil, err
	}
	return toCategoryResponse(created), nil
}

// UpdateCategory 更新税类别
func (uc *TaxUsecase) UpdateCategory(ctx context.Context, req request.UpdateCategoryRequest) (*response.CategoryResponse, error) {
	category := &Category{
		ID:           req.ID,
		Name:         req.Name,
		Rate:         req.Rate,
		DisplayOrder: req.DisplayOrder,
		IsActive:     req.IsActive,
		Description:  req.Description,
	}

	updated, err := uc.categoryRepo.Update(ctx, category)
	if err != nil {
		return nil, err
	}
	return toCategoryResponse(updated), nil
}

// DeleteCategory 删除税类别
func (uc *TaxUsecase) DeleteCategory(ctx context.Context, id uint) error {
	return uc.categoryRepo.Delete(ctx, id)
}

// ==================== 内部转换函数 ====================

// toProviderResponse 领域实体转换为税务提供者响应 DTO
func toProviderResponse(p *Provider) *response.ProviderResponse {
	return &response.ProviderResponse{
		ID:            p.ID,
		Name:          p.Name,
		SystemKeyword: p.SystemKeyword,
		DisplayOrder:  p.DisplayOrder,
		IsActive:      p.IsActive,
		IsPrimary:     p.IsPrimary,
		LogoURL:       p.LogoURL,
		CreatedAt:     p.CreatedAt.Unix(),
		UpdatedAt:     p.UpdatedAt.Unix(),
	}
}

// toCategoryResponse 领域实体转换为税类别响应 DTO
func toCategoryResponse(c *Category) *response.CategoryResponse {
	return &response.CategoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		Rate:         c.Rate,
		DisplayOrder: c.DisplayOrder,
		IsActive:     c.IsActive,
		Description:  c.Description,
		CreatedAt:    c.CreatedAt.Unix(),
		UpdatedAt:    c.UpdatedAt.Unix(),
	}
}