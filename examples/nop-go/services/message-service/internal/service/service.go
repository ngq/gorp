// Package service 服务层。
//
// 服务层是业务逻辑与 HTTP handler 之间的桥梁，负责：
// 1. 组装 biz 层用例和 data 层仓储
// 2. 定义请求/响应 DTO（数据传输对象）
// 3. 将领域实体转换为响应格式
package service

import (
	"context"

	"nop-go/services/message-service/internal/biz"
	"nop-go/services/message-service/internal/data"

	"gorm.io/gorm"
)

// Services 消息服务集合。
//
// 聚合所有子服务，供路由注册时统一注入。
// 新增子服务时在此结构体中添加字段即可。
type Services struct {
	Message *MessageService
}

// NewServices 创建消息服务集合。
//
// 依赖链组装顺序：data（仓储）→ biz（用例）→ service（服务）。
// 此函数是 Wire Provider 的入口，wire_gen.go 会调用它。
func NewServices(db *gorm.DB) *Services {
	// 组装依赖链：仓储 → 用例 → 服务
	tplRepo := data.NewMessageTemplateRepo(db)
	tplUC := biz.NewMessageTemplateUseCase(tplRepo)
	return &Services{
		Message: &MessageService{uc: tplUC},
	}
}

// MessageService 消息服务。
//
// 封装消息模板的业务操作，提供面向 handler 层的调用接口。
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
//
// service 层的响应 DTO，时间字段为格式化字符串。
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

// toMessageTemplateResponse 将领域实体转换为响应结构体。
//
// 时间字段统一格式化为 "2006-01-02 15:04:05" 字符串输出。
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
//
// 调用用例层查询，将领域实体列表转换为响应 DTO 列表。
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
