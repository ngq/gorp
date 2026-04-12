// Package pb 通知服务 gRPC 接口定义
package notification

import (
	"context"
)

// NotificationServiceServer 通知服务服务端接口
type NotificationServiceServer interface {
	// SendEmail 发送邮件
	SendEmail(ctx context.Context, req *SendEmailRequest) (*SendEmailResponse, error)

	// SendSMS 发送短信
	SendSMS(ctx context.Context, req *SendSMSRequest) (*SendSMSResponse, error)

	// SendOrderNotification 发送订单通知
	SendOrderNotification(ctx context.Context, req *SendOrderNotificationRequest) (*SendOrderNotificationResponse, error)

	// SendPaymentNotification 发送支付通知
	SendPaymentNotification(ctx context.Context, req *SendPaymentNotificationRequest) (*SendPaymentNotificationResponse, error)
}

// SendEmailRequest 发送邮件请求
type SendEmailRequest struct {
	To      string
	Subject string
	Content string
	IsHTML  bool
}

// SendEmailResponse 发送邮件响应
type SendEmailResponse struct {
	Success      bool
	MessageID    string
	ErrorMessage string
}

// SendSMSRequest 发送短信请求
type SendSMSRequest struct {
	Phone          string
	Content        string
	TemplateCode   string
	TemplateParams map[string]string
}

// SendSMSResponse 发送短信响应
type SendSMSResponse struct {
	Success      bool
	MessageID    string
	ErrorMessage string
}

// SendOrderNotificationRequest 发送订单通知请求
type SendOrderNotificationRequest struct {
	CustomerID        uint64
	OrderID           uint64
	NotificationType  string // created, paid, shipped, delivered, cancelled
	ExtraData         map[string]string
}

// SendOrderNotificationResponse 发送订单通知响应
type SendOrderNotificationResponse struct {
	Success      bool
	ErrorMessage string
}

// SendPaymentNotificationRequest 发送支付通知请求
type SendPaymentNotificationRequest struct {
	CustomerID        uint64
	PaymentID         uint64
	NotificationType  string // created, success, failed, refunded
	ExtraData         map[string]string
}

// SendPaymentNotificationResponse 发送支付通知响应
type SendPaymentNotificationResponse struct {
	Success      bool
	ErrorMessage string
}