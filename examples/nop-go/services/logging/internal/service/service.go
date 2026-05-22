package service

import (
	"context"

	"nop-go/services/logging/internal/biz"
	"nop-go/services/logging/internal/data"

	"gorm.io/gorm"
)

type Services struct {
	Logging *LoggingService
}

func NewServices(db *gorm.DB) *Services {
	loggingRepo := data.NewLoggingRepo(db)
	loggingUC := biz.NewLoggingUseCase(loggingRepo)
	return &Services{
		Logging: &LoggingService{uc: loggingUC},
	}
}

type LoggingService struct {
	uc *biz.LoggingUseCase
}

type CreateLoggingRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type LoggingResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

func (s *LoggingService) List(ctx context.Context, page, size int) ([]LoggingResponse, int64, error) {
	loggings, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]LoggingResponse, len(loggings))
	for i, u := range loggings {
		items[i] = LoggingResponse{ID: u.ID, Username: u.Username, Email: u.Email}
	}
	return items, total, nil
}

func (s *LoggingService) GetByID(ctx context.Context, id uint) (*LoggingResponse, error) {
	logging, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &LoggingResponse{ID: logging.ID, Username: logging.Username, Email: logging.Email}, nil
}

func (s *LoggingService) Create(ctx context.Context, req CreateLoggingRequest) (*LoggingResponse, error) {
	logging, err := s.uc.Create(ctx, req.Username, req.Email)
	if err != nil {
		return nil, err
	}
	return &LoggingResponse{ID: logging.ID, Username: logging.Username, Email: logging.Email}, nil
}

func (s *LoggingService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}
