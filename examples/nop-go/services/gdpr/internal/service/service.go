package service

import (
	"context"

	"nop-go/services/gdpr/internal/biz"
	"nop-go/services/gdpr/internal/data"

	"gorm.io/gorm"
)

type Services struct {
	Gdpr *GdprService
}

func NewServices(db *gorm.DB) *Services {
	gdprRepo := data.NewGdprRepo(db)
	gdprUC := biz.NewGdprUseCase(gdprRepo)
	return &Services{
		Gdpr: &GdprService{uc: gdprUC},
	}
}

type GdprService struct {
	uc *biz.GdprUseCase
}

type CreateGdprRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type GdprResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (s *GdprService) List(ctx context.Context, page, size int) ([]GdprResponse, int64, error) {
	gdprs, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]GdprResponse, len(gdprs))
	for i, u := range gdprs {
		items[i] = GdprResponse{ID: u.ID, Username: u.Username, Email: u.Email}
	}
	return items, total, nil
}

func (s *GdprService) GetByID(ctx context.Context, id uint) (*GdprResponse, error) {
	gdpr, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GdprResponse{ID: gdpr.ID, Username: gdpr.Username, Email: gdpr.Email}, nil
}

func (s *GdprService) Create(ctx context.Context, req CreateGdprRequest) (*GdprResponse, error) {
	gdpr, err := s.uc.Create(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return &GdprResponse{ID: gdpr.ID, Username: gdpr.Username, Email: gdpr.Email}, nil
}

func (s *GdprService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
