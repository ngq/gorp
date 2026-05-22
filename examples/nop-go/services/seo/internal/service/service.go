package service

import (
	"context"

	"nop-go/services/seo/internal/biz"
	"nop-go/services/seo/internal/data"

	"gorm.io/gorm"
)

type Services struct {
	Seo *SeoService
}

func NewServices(db *gorm.DB) *Services {
	seoRepo := data.NewSeoRepo(db)
	seoUC := biz.NewSeoUseCase(seoRepo)
	return &Services{
		Seo: &SeoService{uc: seoUC},
	}
}

type SeoService struct {
	uc *biz.SeoUseCase
}

type CreateSeoRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type SeoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (s *SeoService) List(ctx context.Context, page, size int) ([]SeoResponse, int64, error) {
	seos, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]SeoResponse, len(seos))
	for i, u := range seos {
		items[i] = SeoResponse{ID: u.ID, Username: u.Username, Email: u.Email}
	}
	return items, total, nil
}

func (s *SeoService) GetByID(ctx context.Context, id uint) (*SeoResponse, error) {
	seo, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SeoResponse{ID: seo.ID, Username: seo.Username, Email: seo.Email}, nil
}

func (s *SeoService) Create(ctx context.Context, req CreateSeoRequest) (*SeoResponse, error) {
	seo, err := s.uc.Create(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return &SeoResponse{ID: seo.ID, Username: seo.Username, Email: seo.Email}, nil
}

func (s *SeoService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
