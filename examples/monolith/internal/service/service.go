// Package service 负责业务编排。
package service

import (
	"context"

	"monolith/internal/biz"
)

// Services 聚合应用服务。
type Services struct {
	Demo *DemoService
}

// NewServices 创建应用服务集合。
func NewServices(biz *biz.Biz) *Services {
	return &Services{
		Demo: &DemoService{uc: biz.Demo},
	}
}

// DemoService 编排 Demo 用例。
type DemoService struct {
	uc *biz.DemoUseCase
}

func (s *DemoService) Create(ctx context.Context, name string) (*biz.Demo, error) {
	return s.uc.Create(ctx, name)
}

func (s *DemoService) GetByID(ctx context.Context, id uint) (*biz.Demo, error) {
	return s.uc.GetByID(ctx, id)
}

func (s *DemoService) List(ctx context.Context, page, pageSize int) ([]*biz.Demo, int64, error) {
	return s.uc.List(ctx, page, pageSize)
}

func (s *DemoService) Update(ctx context.Context, id uint, name string) (*biz.Demo, error) {
	return s.uc.Update(ctx, id, name)
}

func (s *DemoService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
