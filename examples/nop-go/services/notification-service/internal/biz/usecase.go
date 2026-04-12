// Package biz 通知服务业务逻辑层
package biz

import (
	"bytes"
	"context"
	"text/template"

	"nop-go/services/notification-service/internal/data"
	"nop-go/services/notification-service/internal/models"
)

type NotificationUseCase struct {
	notificationRepo data.NotificationRepository
	templateRepo     data.NotificationTemplateRepository
	smsRepo          data.SMSRecordRepository
}

func NewNotificationUseCase(
	notificationRepo data.NotificationRepository,
	templateRepo data.NotificationTemplateRepository,
	smsRepo data.SMSRecordRepository,
) *NotificationUseCase {
	return &NotificationUseCase{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
		smsRepo:          smsRepo,
	}
}

func (uc *NotificationUseCase) SendEmail(ctx context.Context, req *models.SendEmailRequest) error {
	notification := &models.Notification{
		Type:      "email",
		Recipient: req.To,
		Subject:   req.Subject,
		Content:   req.Content,
		Status:    "pending",
	}

	return uc.notificationRepo.Create(ctx, notification)
}

func (uc *NotificationUseCase) SendSMS(ctx context.Context, req *models.SendSMSRequest) error {
	notification := &models.Notification{
		Type:      "sms",
		Recipient: req.Phone,
		Content:   req.Content,
		Status:    "pending",
	}

	return uc.notificationRepo.Create(ctx, notification)
}

func (uc *NotificationUseCase) SendTemplate(ctx context.Context, req *models.SendTemplateRequest) error {
	tmpl, err := uc.templateRepo.GetByCode(ctx, req.TemplateCode)
	if err != nil {
		return err
	}

	var content bytes.Buffer
	t, err := template.New("notification").Parse(tmpl.Content)
	if err != nil {
		return err
	}

	if err := t.Execute(&content, req.Variables); err != nil {
		return err
	}

	notification := &models.Notification{
		Type:      tmpl.Type,
		Recipient: req.Recipient,
		Subject:   tmpl.Subject,
		Content:   content.String(),
		Status:    "pending",
	}

	return uc.notificationRepo.Create(ctx, notification)
}

func (uc *NotificationUseCase) GetPendingNotifications(ctx context.Context, limit int) ([]*models.Notification, error) {
	return uc.notificationRepo.GetPending(ctx, limit)
}

func (uc *NotificationUseCase) GetNotification(ctx context.Context, id uint64) (*models.Notification, error) {
	return uc.notificationRepo.GetByID(ctx, id)
}

func (uc *NotificationUseCase) MarkAsSent(ctx context.Context, id uint64) error {
	notification, err := uc.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	notification.Status = "sent"
	return uc.notificationRepo.Update(ctx, notification)
}

func (uc *NotificationUseCase) MarkAsFailed(ctx context.Context, id uint64, errorMessage string) error {
	notification, err := uc.notificationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	notification.Status = "failed"
	notification.ErrorMessage = errorMessage
	notification.RetryCount++

	return uc.notificationRepo.Update(ctx, notification)
}

type TemplateUseCase struct {
	templateRepo data.NotificationTemplateRepository
}

func NewTemplateUseCase(templateRepo data.NotificationTemplateRepository) *TemplateUseCase {
	return &TemplateUseCase{templateRepo: templateRepo}
}

func (uc *TemplateUseCase) CreateTemplate(ctx context.Context, tmpl *models.NotificationTemplate) error {
	return uc.templateRepo.Create(ctx, tmpl)
}

func (uc *TemplateUseCase) GetTemplate(ctx context.Context, code string) (*models.NotificationTemplate, error) {
	return uc.templateRepo.GetByCode(ctx, code)
}

func (uc *TemplateUseCase) ListTemplates(ctx context.Context) ([]*models.NotificationTemplate, error) {
	return uc.templateRepo.List(ctx)
}

func (uc *TemplateUseCase) UpdateTemplate(ctx context.Context, tmpl *models.NotificationTemplate) error {
	return uc.templateRepo.Update(ctx, tmpl)
}

func (uc *TemplateUseCase) DeleteTemplate(ctx context.Context, id uint64) error {
	return uc.templateRepo.Delete(ctx, id)
}