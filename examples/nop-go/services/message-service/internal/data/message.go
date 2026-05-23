// Package data 数据访问层。
//
// 负责领域实体与数据库之间的映射和持久化操作。
// data 层实现 biz 层定义的仓储接口，完成依赖倒置。
package data

import (
	"context"
	"time"

	"nop-go/services/message-service/internal/biz"

	"gorm.io/gorm"
)

// MessageTemplatePO 消息模板持久化对象。
//
// PO（Persistent Object）与数据库表一一对应，包含 gorm/db/json 标签。
// 与领域实体 MessageTemplate 分离，避免持久化细节泄漏到业务层。
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
//
// 指定 GORM 操作的数据库表名为 message_templates。
func (MessageTemplatePO) TableName() string {
	return "message_templates"
}

// ToEntity 转换为领域实体。
//
// 将持久化对象转换为业务层使用的领域实体，剥离持久化标签。
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
//
// 实现 biz.MessageTemplateRepository 接口，通过 GORM 操作数据库。
type messageTemplateRepo struct {
	db *gorm.DB
}

// NewMessageTemplateRepo 创建消息模板仓储。
//
// 返回 biz 层定义的仓储接口，实现依赖倒置。
func NewMessageTemplateRepo(db *gorm.DB) biz.MessageTemplateRepository {
	return &messageTemplateRepo{db: db}
}

// Create 创建消息模板。
//
// 将领域实体转换为 PO 后通过 GORM 写入数据库。
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
//
// 从数据库查询 PO 后转换为领域实体返回。
func (r *messageTemplateRepo) GetByID(ctx context.Context, id uint) (*biz.MessageTemplate, error) {
	var po MessageTemplatePO
	if err := r.db.WithContext(ctx).First(&po, id).Error; err != nil {
		return nil, err
	}
	return po.ToEntity(), nil
}

// List 获取消息模板列表。
//
// 支持分页查询，先统计总数再查询分页数据。
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
//
// 使用 map 方式更新指定字段，避免零值覆盖问题。
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
