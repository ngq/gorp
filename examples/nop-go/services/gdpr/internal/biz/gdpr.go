// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Gdpr Gdpr领域实体。
type Gdpr struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GdprRepository Gdpr仓储接口。
type GdprRepository interface {
	Create(ctx context.Context, gdpr *Gdpr) error
	GetByID(ctx context.Context, id uint) (*Gdpr, error)
	GetByUsername(ctx context.Context, username string) (*Gdpr, error)
	List(ctx context.Context, page, size int) ([]*Gdpr, int64, error)
	Delete(ctx context.Context, id uint) error
}

// GdprUseCase Gdpr用例。
type GdprUseCase struct {
	repo GdprRepository
}

// NewGdprUseCase 创建Gdpr用例。
func NewGdprUseCase(repo GdprRepository) *GdprUseCase {
	return &GdprUseCase{repo: repo}
}

// Create 创建Gdpr。
func (uc *GdprUseCase) Create(ctx context.Context, username, email string) (*Gdpr, error) {
	gdpr := &Gdpr{
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, gdpr); err != nil {
		return nil, err
	}
	return gdpr, nil
}

// GetByID 根据ID获取Gdpr。
func (uc *GdprUseCase) GetByID(ctx context.Context, id uint) (*Gdpr, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取Gdpr列表。
func (uc *GdprUseCase) List(ctx context.Context, page, size int) ([]*Gdpr, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除Gdpr。
func (uc *GdprUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}