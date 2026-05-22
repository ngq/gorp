// Package biz 业务逻辑层。
package biz

import (
	"context"
	"time"
)

// Logging Logging领域实体。
type Logging struct {
	ID        uint
	Username  string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LoggingRepository Logging仓储接口。
type LoggingRepository interface {
	Create(ctx context.Context, logging *Logging) error
	GetByID(ctx context.Context, id uint) (*Logging, error)
	GetByUsername(ctx context.Context, username string) (*Logging, error)
	List(ctx context.Context, page, size int) ([]*Logging, int64, error)
	Delete(ctx context.Context, id uint) error
}

// LoggingUseCase Logging用例。
type LoggingUseCase struct {
	repo LoggingRepository
}

// NewLoggingUseCase 创建Logging用例。
func NewLoggingUseCase(repo LoggingRepository) *LoggingUseCase {
	return &LoggingUseCase{repo: repo}
}

// Create 创建Logging。
func (uc *LoggingUseCase) Create(ctx context.Context, username, email string) (*Logging, error) {
	logging := &Logging{
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := uc.repo.Create(ctx, logging); err != nil {
		return nil, err
	}
	return logging, nil
}

// GetByID 根据ID获取Logging。
func (uc *LoggingUseCase) GetByID(ctx context.Context, id uint) (*Logging, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取Logging列表。
func (uc *LoggingUseCase) List(ctx context.Context, page, size int) ([]*Logging, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Delete 删除Logging。
func (uc *LoggingUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}