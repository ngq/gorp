// Package biz 业务逻辑层。
package biz

import (
	"context"
	"fmt"
	"time"
)

// MessageTemplate 消息模板领域实体。
type MessageTemplate struct {
	ID           uint      // 模板ID
	Name         string    // 模板名称
	Subject      string    // 邮件主题
	Body         string    // 邮件正文
	EmailAccount string    // 发件邮箱账号
	IsActive     bool      // 是否启用
	CreatedAt    time.Time // 创建时间
	UpdatedAt    time.Time // 更新时间
}

// MessageTemplateRepository 消息模板仓储接口。
type MessageTemplateRepository interface {
	// Create 创建消息模板
	Create(ctx context.Context, tpl *MessageTemplate) error
	// GetByID 根据ID获取消息模板
	GetByID(ctx context.Context, id uint) (*MessageTemplate, error)
	// List 获取消息模板列表
	List(ctx context.Context, page, size int) ([]*MessageTemplate, int64, error)
	// Update 更新消息模板
	Update(ctx context.Context, tpl *MessageTemplate) error
	// Delete 删除消息模板
	Delete(ctx context.Context, id uint) error
}

// MessageTemplateUseCase 消息模板用例。
type MessageTemplateUseCase struct {
	repo MessageTemplateRepository
}

// NewMessageTemplateUseCase 创建消息模板用例。
func NewMessageTemplateUseCase(repo MessageTemplateRepository) *MessageTemplateUseCase {
	return &MessageTemplateUseCase{repo: repo}
}

// Create 创建消息模板。
func (uc *MessageTemplateUseCase) Create(ctx context.Context, name, subject, body, emailAccount string, isActive bool) (*MessageTemplate, error) {
	tpl := &MessageTemplate{
		Name:         name,
		Subject:      subject,
		Body:         body,
		EmailAccount: emailAccount,
		IsActive:     isActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.Create(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

// GetByID 根据ID获取消息模板。
func (uc *MessageTemplateUseCase) GetByID(ctx context.Context, id uint) (*MessageTemplate, error) {
	return uc.repo.GetByID(ctx, id)
}

// List 获取消息模板列表。
func (uc *MessageTemplateUseCase) List(ctx context.Context, page, size int) ([]*MessageTemplate, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新消息模板。
func (uc *MessageTemplateUseCase) Update(ctx context.Context, id uint, name, subject, body, emailAccount string, isActive bool) (*MessageTemplate, error) {
	tpl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tpl.Name = name
	tpl.Subject = subject
	tpl.Body = body
	tpl.EmailAccount = emailAccount
	tpl.IsActive = isActive
	tpl.UpdatedAt = time.Now()
	if err := uc.repo.Update(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

// Delete 删除消息模板。
func (uc *MessageTemplateUseCase) Delete(ctx context.Context, id uint) error {
	return uc.repo.Delete(ctx, id)
}

// Test 测试消息模板，模拟发送测试邮件。
func (uc *MessageTemplateUseCase) Test(ctx context.Context, id uint, toEmail string) error {
	tpl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	// 模拟发送测试邮件，实际项目中应调用邮件服务
	_ = fmt.Sprintf("发送测试邮件到 %s，主题: %s", toEmail, tpl.Subject)
	return nil
}

// Copy 复制消息模板。
func (uc *MessageTemplateUseCase) Copy(ctx context.Context, id uint) (*MessageTemplate, error) {
	tpl, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	newTpl := &MessageTemplate{
		Name:         tpl.Name + " (副本)",
		Subject:      tpl.Subject,
		Body:         tpl.Body,
		EmailAccount: tpl.EmailAccount,
		IsActive:     false, // 复制后默认不启用
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := uc.repo.Create(ctx, newTpl); err != nil {
		return nil, err
	}
	return newTpl, nil
}
