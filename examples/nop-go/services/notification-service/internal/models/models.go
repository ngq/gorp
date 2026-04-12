// Package models 通知服务数据模型
package models

import (
	"time"
)

// Notification 通知记录
type Notification struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Type         string    `gorm:"size:16;not null;index" json:"type"` // email, sms, push
	Recipient    string    `gorm:"size:128;not null;index" json:"recipient"`
	Subject      string    `gorm:"size:256" json:"subject"`
	Content      string    `gorm:"type:text;not null" json:"content"`
	Status       string    `gorm:"size:16;default:'pending';index" json:"status"` // pending, sent, failed
	RetryCount   int       `gorm:"default:0" json:"retry_count"`
	ErrorMessage string    `gorm:"size:512" json:"error_message"`
	SentAt       *time.Time `json:"sent_at"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationTemplate 通知模板
type NotificationTemplate struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:64;not null" json:"name"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Type      string    `gorm:"size:16;not null" json:"type"` // email, sms, push
	Subject   string    `gorm:"size:256" json:"subject"`      // 邮件主题
	Content   string    `gorm:"type:text;not null" json:"content"` // 支持 {{.Variable}} 模板
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// EmailQueue 邮件队列
type EmailQueue struct {
	ID          uint64     `gorm:"primaryKey" json:"id"`
	To          string     `gorm:"size:128;not null" json:"to"`
	Subject     string     `gorm:"size:256;not null" json:"subject"`
	Body        string     `gorm:"type:text;not null" json:"body"`
	Status      string     `gorm:"size:16;default:'pending'" json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

func (EmailQueue) TableName() string {
	return "email_queue"
}

// SMSRecord 短信记录
type SMSRecord struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"size:32;not null;index" json:"phone"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Code      string    `gorm:"size:8" json:"code"` // 验证码
	Status    string    `gorm:"size:16;default:'pending'" json:"status"`
	SentAt    *time.Time `json:"sent_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (SMSRecord) TableName() string {
	return "sms_records"
}

// DTO
type SendEmailRequest struct {
	To      string `json:"to" binding:"required,email"`
	Subject string `json:"subject" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type SendSMSRequest struct {
	Phone   string `json:"phone" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type SendTemplateRequest struct {
	TemplateCode string                 `json:"template_code" binding:"required"`
	Recipient    string                 `json:"recipient" binding:"required"`
	Variables    map[string]interface{} `json:"variables"`
}

type NotificationResponse struct {
	ID        uint64 `json:"id"`
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func ToNotificationResponse(n *Notification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		Recipient: n.Recipient,
		Status:    n.Status,
		CreatedAt: n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}