package service

import (
	"context"

	"nop-go/services/plugin/internal/biz"
	"nop-go/services/plugin/internal/data"

	"gorm.io/gorm"
)

type Services struct {
	Plugin *PluginService
}

func NewServices(db *gorm.DB) *Services {
	pluginRepo := data.NewPluginRepo(db)
	pluginUC := biz.NewPluginUseCase(pluginRepo)
	return &Services{
		Plugin: &PluginService{uc: pluginUC},
	}
}

type PluginService struct {
	uc *biz.PluginUseCase
}

type CreatePluginRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type PluginResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (s *PluginService) List(ctx context.Context, page, size int) ([]PluginResponse, int64, error) {
	plugins, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]PluginResponse, len(plugins))
	for i, u := range plugins {
		items[i] = PluginResponse{ID: u.ID, Username: u.Username, Email: u.Email}
	}
	return items, total, nil
}

func (s *PluginService) GetByID(ctx context.Context, id uint) (*PluginResponse, error) {
	plugin, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PluginResponse{ID: plugin.ID, Username: plugin.Username, Email: plugin.Email}, nil
}

func (s *PluginService) Create(ctx context.Context, req CreatePluginRequest) (*PluginResponse, error) {
	plugin, err := s.uc.Create(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return &PluginResponse{ID: plugin.ID, Username: plugin.Username, Email: plugin.Email}, nil
}

func (s *PluginService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
