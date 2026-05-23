// Package service 定义 message-service 的 gRPC RPC 接口。
// gorp proto from-service 命令通过解析此文件中的接口定义来生成 .proto 文件。
// 只有其他服务需要通过 gRPC 跨服务调用的方法才定义在此接口中。
package service

import "context"

// MessageRPC 消息服务 gRPC 接口 —— 定义其他服务需要跨服务调用的方法。
// 任何服务需要发送通知邮件时，可通过 GetMessageTemplate 获取模板内容。
type MessageRPC interface {
	// GetMessageTemplate 根据 ID 获取消息模板（其他服务发送通知时获取模板内容）
	GetMessageTemplate(ctx context.Context, req *GetMessageTemplateReq) (*GetMessageTemplateResp, error)
}

// ======================== gRPC 请求/响应类型 ========================

// GetMessageTemplateReq 获取消息模板请求
type GetMessageTemplateReq struct {
	ID uint32 `json:"id" remark:"模板ID"`
}

// GetMessageTemplateResp 获取消息模板响应
type GetMessageTemplateResp struct {
	ID           uint32 `json:"id" remark:"模板ID"`
	Name         string `json:"name" remark:"模板名称"`
	Subject      string `json:"subject" remark:"邮件主题"`
	Body         string `json:"body" remark:"邮件正文"`
	EmailAccount string `json:"email_account" remark:"发件邮箱账号"`
	IsActive     bool   `json:"is_active" remark:"是否启用"`
}