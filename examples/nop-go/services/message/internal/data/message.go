// Package data 数据访问层。
package data

import (
	"context"
	"time"

	"nop-go/services/message/internal/biz"

	"gorm.io/gorm"
)

// MessageTemplatePO 消息模板持久化对象。
type MessageTemplatePO struct {
	ID           uint      `gorm:"primaryKey;column:id" db:"id" json:"id"`
	Name         string    `gorm:"size:256;column:name" db:"name" json:"name"`
	Subject      string    `gorm:"size:512;column:subject" db:"subject" json:"subject"`
	Body         string    `gorm:"type:text;column:body" db:"body" json:"body"`
	EmailAccount string    `gorm:"size:256;column:email_account" db:"email_account" json:"email_account"`
	IsActive     bool      `gorm:"column:is_active" db:"is_active" json:"is_active"`
	CreatedAt    time.Time `gorm:"autoCreateTime;column:created_at" db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime;column:updated_at" db:"updated_at" json:"updated_at"`
}

// TableName 表名。
func (MessageTemplatePO) TableName() string {
	return "message_templates"
}

// ToEntity 转换为领域实体。
func (po *MessageTemplatePO) ToEntity() *biz.MessageTemplate {
	return &biz.MessageTemplate{
		ID:           po.ID,
		Name:         po.Name,
		Subject:      po.Subject,
		Body:         po.Body,
		EmailAccount: po.EmailAccount,
		IsActive:     po.IsActive,
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// messageTemplateRepo 消息模板仓储实现。
type messageTemplateRepo struct {
	db *gorm.DB
}

// NewMessageTemplateRepo 创建消息模板仓储。
func NewMessageTemplateRepo(db *gorm.DB) biz.MessageTemplateRepository {
	return &messageTemplateRepo{db: db}
}

// Create 创建消息模板。
func (r *messageTemplateRepo) Create(ctx context.Context, tpl *biz.MessageTemplate) error {
	po := &MessageTemplatePO{
		Name:         tpl.Name,
		Subject:      tpl.Subject,
		Body:         tpl.Body,
		EmailAccount: tpl.EmailAccount,
		IsActive:     tpl.IsActive,
		CreatedAt:    tpl.CreatedAt,
		UpdatedAt:    tpl.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID获取消息模板。
func (r *messageTemplateRepo) GetByID(ctx context.Context, id uint) (*biz.MessageTemplate, error) {
	var po MessageTemplatePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取消息模板列表。
func (r *messageTemplateRepo) List(ctx context.Context, page, size int) ([]*biz.MessageTemplate, int64, error) {
	var pos []MessageTemplatePO
	var total int64

	r.db.WithContext(ctx).Model(&MessageTemplatePO{}).Count(&total)

	offset := (page - 1) * size
	if err := r.db.WithContext(ctx).Offset(offset).Limit(size).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	tpls := make([]*biz.MessageTemplate, len(pos))
	for i, po := range pos {
		tpls[i] = po.ToEntity()
	}

	return tpls, total, nil
}

// Update 更新消息模板。
func (r *messageTemplateRepo) Update(ctx context.Context, tpl *biz.MessageTemplate) error {
	return r.db.WithContext(ctx).Model(&MessageTemplatePO{}).Where("id = ?", tpl.ID).Updates(map[string]interface{}{
		"name":          tpl.Name,
		"subject":       tpl.Subject,
		"body":          tpl.Body,
		"email_account": tpl.EmailAccount,
		"is_active":     tpl.IsActive,
		"updated_at":    tpl.UpdatedAt,
	}).Error
}

// Delete 删除消息模板。
func (r *messageTemplateRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&MessageTemplatePO{}, id).Error
}
