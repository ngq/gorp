// Package biz 业务逻辑层。
// 定义 SEO 服务的领域实体、仓储接口和用例。
// 包含 SEO 元数据的创建、查询、删除等核心能力。
package biz

import (
	"context"
	"time"
)

// SeoMeta SEO 元数据领域实体。
type SeoMeta struct {
	ID        uint      // SEO元数据ID
	Username  string    // 关联用户名
	Email     string    // 关联邮箱
	CreatedAt time.Time // 创建时间
	UpdatedAt time.Time // 更新时间
}

// SeoMetaRepository SEO 元数据仓储接口。
type SeoMetaRepository interface {
	Create(ctx context.Context, seo *SeoMeta) error
	GetByID(ctx context.Context, id uint) (*SeoMeta, error)
	GetByUsername(ctx context.Context, username string) (*SeoMeta, error)
	List(ctx context.Context, page, size int) ([]*SeoMeta, int64, error)
	Delete(ctx context.Context, id uint) error
}

// SeoUseCase SEO 用例。
type SeoUseCase struct {
	repo SeoMetaRepository
}

// NewSeoUseCase 创建 SEO 用例。
func NewSeoUseCase(repo SeoMetaRepository) *SeoUseCase {
	return &SeoUseCase{repo: repo}
}

// Create 创建 SEO 元数据。
func (uc *SeoUseCase) Create(ctx context.Context, username, email string) (*SeoMeta, error) {
	seo := &SeoMeta{
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, seo); err != nil {
		return nil, err
	}
	return seo, nil
}

// GetByID 根据ID获取 SEO 元数据。
func (uc *SeoUseCase) GetByID(ctx context.Context, id uint) (*SeoMeta, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取 SEO 元数据列表。
func (uc *SeoUseCase) List(ctx context.Context, page, size int) ([]*SeoMeta, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除 SEO 元数据。
func (uc *SeoUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}
