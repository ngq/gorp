// Package biz 业务逻辑层
//
// 中文说明：
// - 核心业务逻辑，领域对象，领域服务；
// - 定义 Repository 接口，由 data 层实现；
// - 不依赖外部框架（Gin、GORM 等），保持纯粹。
package biz

import (
	"context"
	"time"
)

// Demo 示例领域实体。
type Demo struct {
	ID        uint
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DemoRepository Demo 仓储接口。
//
// 中文说明：
// - 定义数据访问契约，由 data 层实现；
// - biz 层只依赖接口，不依赖具体实现。
type DemoRepository interface {
	Create(ctx context.Context, demo *Demo) error
	GetByID(ctx context.Context, id uint) (*Demo, error)
	List(ctx context.Context, page, pageSize int) ([]*Demo, int64, error)
	Update(ctx context.Context, demo *Demo) error
	Delete(ctx context.Context, id uint) error
}

// Biz 业务层聚合。
type Biz struct {
	Demo *DemoUseCase
}

// NewBiz 创建业务层。
func NewBiz(repo DemoRepository) *Biz {
	return &Biz{
		Demo: NewDemoUseCase(repo),
	}
}

// DemoUseCase Demo 用例。
type DemoUseCase struct {
	repo DemoRepository
}

// NewDemoUseCase 创建 Demo 用例。
func NewDemoUseCase(repo DemoRepository) *DemoUseCase {
	return &DemoUseCase{repo: repo}
}

// Create 创建 Demo。
func (uc *DemoUseCase) Create(ctx context.Context, name string) (*Demo, error) {
	demo := &Demo{Name: name}
	if err := uc.repo.Create(ctx, demo); err != nil {
		return nil, err
	}
	return demo, nil
}

// GetByID 根据 ID 获取 Demo。
func (uc *DemoUseCase) GetByID(ctx context.Context, id uint) (*Demo, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取 Demo 列表。
func (uc *DemoUseCase) List(ctx context.Context, page, pageSize int) ([]*Demo, int64, error) {
	return uc.repo.List(ctx, page, pageSize)
}

// Update 更新 Demo。
func (uc *DemoUseCase) Update(ctx context.Context, id uint, name string) (*Demo, error) {
	demo, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	demo.Name = name
	if err := uc.repo.Update(ctx, demo); err != nil {
		return nil, err
	}
	return demo, nil
}

// Delete 删除 Demo。
func (uc *DemoUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}