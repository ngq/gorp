package service

import (
	"context"

	"nop-go/services/message/internal/biz"
	"nop-go/services/message/internal/data"

	"gorm.io/gorm"
)

// Services 消息服务集合。
type Services struct {
	Message *MessageService
}

// NewServices 创建消息服务集合。
func NewServices(db *gorm.DB) *Services {
	tplRepo := data.NewMessageTemplateRepo(db)
	tplUC := biz.NewMessageTemplateUseCase(tplRepo)
	return &Services{
		Message: &MessageService{uc: tplUC},
	}
}

// MessageService 消息服务。
type MessageService struct {
	uc *biz.MessageTemplateUseCase
}

// CreateMessageTemplateRequest 创建消息模板请求。
type CreateMessageTemplateRequest struct {
	Name         string `json:"name" binding:"required"`          // 模板名称
	Subject      string `json:"subject" binding:"required"`       // 邮件主题
	Body         string `json:"body" binding:"required"`          // 邮件正文
	EmailAccount string `json:"email_account" binding:"required"` // 发件邮箱账号
	IsActive     bool   `json:"is_active"`                        // 是否启用
}

// UpdateMessageTemplateRequest 更新消息模板请求。
type UpdateMessageTemplateRequest struct {
	Name         string `json:"name" binding:"required"`          // 模板名称
	Subject      string `json:"subject" binding:"required"`       // 邮件主题
	Body         string `json:"body" binding:"required"`          // 邮件正文
	EmailAccount string `json:"email_account" binding:"required"` // 发件邮箱账号
	IsActive     bool   `json:"is_active"`                        // 是否启用
}

// MessageTemplateResponse 消息模板响应。
type MessageTemplateResponse struct {
	ID           uint   `json:"id"`            // 模板ID
	Name         string `json:"name"`          // 模板名称
	Subject      string `json:"subject"`       // 邮件主题
	Body         string `json:"body"`          // 邮件正文
	EmailAccount string `json:"email_account"` // 发件邮箱账号
	IsActive     bool   `json:"is_active"`     // 是否启用
	CreatedAt    string `json:"created_at"`    // 创建时间
	UpdatedAt    string `json:"updated_at"`    // 更新时间
}

// toResponse 将领域实体转换为响应结构体。
func toMessageTemplateResponse(tpl *biz.MessageTemplate) *MessageTemplateResponse {
	return &MessageTemplateResponse{
		ID:           tpl.ID,
		Name:         tpl.Name,
		Subject:      tpl.Subject,
		Body:         tpl.Body,
		EmailAccount: tpl.EmailAccount,
		IsActive:     tpl.IsActive,
		CreatedAt:    tpl.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    tpl.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// List 获取消息模板列表。
func (s *MessageService) List(ctx context.Context, page, size int) ([]MessageTemplateResponse, int64, error) {
	tpls, total, err := s.uc.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	items := make([]MessageTemplateResponse, len(tpls))
	for i, tpl := range tpls {
		items[i] = *toMessageTemplateResponse(tpl)
	}
	return items, total, nil
}

// GetByID 根据ID获取消息模板。
func (s *MessageService) GetByID(ctx context.Context, id uint) (*MessageTemplateResponse, error) {
	tpl, err := s.uc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toMessageTemplateResponse(tpl), nil
}

// Create 创建消息模板。
func (s *MessageService) Create(ctx context.Context, req CreateMessageTemplateRequest) (*MessageTemplateResponse, error) {
	tpl, err := s.uc.Create(ctx, req.Name, req.Subject, req.Body, req.EmailAccount, req.IsActive)
	if err != nil {
		return nil, err
	}
	return toMessageTemplateResponse(tpl), nil
}

// Update 更新消息模板。
func (s *MessageService) Update(ctx context.Context, id uint, req UpdateMessageTemplateRequest) (*MessageTemplateResponse, error) {
	tpl, err := s.uc.Update(ctx, id, req.Name, req.Subject, req.Body, req.EmailAccount, req.IsActive)
	if err != nil {
		return nil, err
	}
	return toMessageTemplateResponse(tpl), nil
}

// Delete 删除消息模板。
func (s *MessageService) Delete(ctx context.Context, id uint) error {
	return s.uc.Delete(ctx, id)
}

// Test 测试消息模板。
func (s *MessageService) Test(ctx context.Context, id uint, toEmail string) error {
	return s.uc.Test(ctx, id, toEmail)
}

// Copy 复制消息模板。
func (s *MessageService) Copy(ctx context.Context, id uint) (*MessageTemplateResponse, error) {
	tpl, err := s.uc.Copy(ctx, id)
	if err != nil {
		return nil, err
	}
	return toMessageTemplateResponse(tpl), nil
}
