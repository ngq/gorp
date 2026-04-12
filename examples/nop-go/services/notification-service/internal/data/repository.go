// Package data 通知服务数据访问层
package data

import (
	"context"

	"nop-go/services/notification-service/internal/models"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *models.Notification) error
	GetByID(ctx context.Context, id uint64) (*models.Notification, error)
	GetPending(ctx context.Context, limit int) ([]*models.Notification, error)
	Update(ctx context.Context, notification *models.Notification) error
}

type NotificationTemplateRepository interface {
	Create(ctx context.Context, template *models.NotificationTemplate) error
	GetByCode(ctx context.Context, code string) (*models.NotificationTemplate, error)
	Update(ctx context.Context, template *models.NotificationTemplate) error
	Delete(ctx context.Context, id uint64) error
	List(ctx context.Context) ([]*models.NotificationTemplate, error)
}

type SMSRecordRepository interface {
	Create(ctx context.Context, record *models.SMSRecord) error
	GetByPhoneAndCode(ctx context.Context, phone, code string) (*models.SMSRecord, error)
}

type notificationRepo struct{ db *gorm.DB }

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) Create(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepo) GetByID(ctx context.Context, id uint64) (*models.Notification, error) {
	var n models.Notification
	err := r.db.WithContext(ctx).First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *notificationRepo) GetPending(ctx context.Context, limit int) ([]*models.Notification, error) {
	var list []*models.Notification
	err := r.db.WithContext(ctx).Where("status = ?", "pending").
		Order("created_at ASC").Limit(limit).Find(&list).Error
	return list, err
}

func (r *notificationRepo) Update(ctx context.Context, n *models.Notification) error {
	return r.db.WithContext(ctx).Save(n).Error
}

type notificationTemplateRepo struct{ db *gorm.DB }

func NewNotificationTemplateRepository(db *gorm.DB) NotificationTemplateRepository {
	return &notificationTemplateRepo{db: db}
}

func (r *notificationTemplateRepo) Create(ctx context.Context, t *models.NotificationTemplate) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *notificationTemplateRepo) GetByCode(ctx context.Context, code string) (*models.NotificationTemplate, error) {
	var t models.NotificationTemplate
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *notificationTemplateRepo) Update(ctx context.Context, t *models.NotificationTemplate) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *notificationTemplateRepo) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&models.NotificationTemplate{}, id).Error
}

func (r *notificationTemplateRepo) List(ctx context.Context) ([]*models.NotificationTemplate, error) {
	var list []*models.NotificationTemplate
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&list).Error
	return list, err
}

type smsRecordRepo struct{ db *gorm.DB }

func NewSMSRecordRepository(db *gorm.DB) SMSRecordRepository {
	return &smsRecordRepo{db: db}
}

func (r *smsRecordRepo) Create(ctx context.Context, s *models.SMSRecord) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *smsRecordRepo) GetByPhoneAndCode(ctx context.Context, phone, code string) (*models.SMSRecord, error) {
	var s models.SMSRecord
	err := r.db.WithContext(ctx).Where("phone = ? AND code = ?", phone, code).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}