// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Seo Seo领域实体。
type Seo struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SeoRepository Seo仓储接口。
type SeoRepository interface {
	Create(ctx context.Context, seo *Seo) error
	GetByID(ctx context.Context, id uint) (*Seo, error)
	GetByUsername(ctx context.Context, username string) (*Seo, error)
	List(ctx context.Context, page, size int) ([]*Seo, int64, error)
	Delete(ctx context.Context, id uint) error
}

// SeoUseCase Seo用例。
type SeoUseCase struct {
	repo SeoRepository
}

// NewSeoUseCase 创建Seo用例。
func NewSeoUseCase(repo SeoRepository) *SeoUseCase {
	return &SeoUseCase{repo: repo}
}

// Create 创建Seo。
func (uc *SeoUseCase) Create(ctx context.Context, username, email string) (*Seo, error) {
	seo := &Seo{
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

// GetByID 根据ID获取Seo。
func (uc *SeoUseCase) GetByID(ctx context.Context, id uint) (*Seo, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取Seo列表。
func (uc *SeoUseCase) List(ctx context.Context, page, size int) ([]*Seo, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除Seo。
func (uc *SeoUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}