// Package biz 业务逻辑层。
//
// 定义消息模板的领域实体、仓储接口和用例。
// biz 层不依赖任何外部基础设施，仅通过仓储接口与数据层解耦。
package biz

import (
	"context"
	"fmt"
	"time"
)

// MessageTemplate 消息模板领域实体。
//
// 领域实体只包含业务属性，不包含持久化标签（如 gorm tag）。
// 与数据库的映射由 data 层的 PO（持久化对象）负责。
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
//
// 仓储接口由 biz 层定义，data 层实现，实现依赖倒置。
// 业务层只关心"能做什么"，不关心"怎么做"。
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
//
// 用例封装业务流程，协调领域实体与仓储接口完成业务操作。
// 每个用例方法对应一个完整的业务场景。
type MessageTemplateUseCase struct {
	repo MessageTemplateRepository
}

// NewMessageTemplateUseCase 创建消息模板用例。
//
// 通过构造函数注入仓储依赖，确保用例可测试。
func NewMessageTemplateUseCase(repo MessageTemplateRepository) *MessageTemplateUseCase {
	return &MessageTemplateUseCase{repo: repo}
}

// Create 创建消息模板。
//
// 流程：构造领域实体 → 调用仓储持久化 → 返回创建结果。
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
//
// 支持分页查询，返回模板列表和总数。
func (uc *MessageTemplateUseCase) List(ctx context.Context, page, size int) ([]*MessageTemplate, int64, error) {
	return uc.repo.List(ctx, page, size)
}

// Update 更新消息模板。
//
// 流程：先查询现有模板 → 更新字段 → 持久化。
// 若模板不存在，仓储层会返回错误。
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
//
// 实际项目中应调用邮件服务发送邮件，此处仅做模拟。
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
//
// 流程：查询源模板 → 构造副本（名称加" (副本)"后缀，默认不启用）→ 持久化。
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
