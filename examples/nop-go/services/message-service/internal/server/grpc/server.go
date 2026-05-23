// Package grpc 实现 message-service 的 gRPC 服务端。
// MessageRPCServer 实现 messagev1.UnimplementedMessageRPCServer 接口，
// 将 gRPC 请求委托给已有的 MessageService 处理。
package grpc

import (
	"context"

	messagev1 "nop-go/services/message-service/api/message/v1"
	"nop-go/services/message-service/internal/service"

	"google.golang.org/grpc"
)

// MessageRPCServer 消息服务 gRPC 服务端实现。
type MessageRPCServer struct {
	messagev1.UnimplementedMessageRPCServer
	messageSvc *service.MessageService
}

// NewMessageRPCServer 创建消息服务 gRPC 服务端。
func NewMessageRPCServer(messageSvc *service.MessageService) *MessageRPCServer {
	return &MessageRPCServer{messageSvc: messageSvc}
}

// GetMessageTemplate 根据 ID 获取消息模板 —— 供任何服务发送通知时获取模板内容。
func (s *MessageRPCServer) GetMessageTemplate(ctx context.Context, req *messagev1.GetMessageTemplateReq) (*messagev1.GetMessageTemplateResp, error) {
	tpl, err := s.messageSvc.GetByID(ctx, uint(req.GetId()))
	if err != nil {
		return nil, err
	}
	return &messagev1.GetMessageTemplateResp{
		Id:           uint32(tpl.ID),
		Name:         tpl.Name,
		Subject:      tpl.Subject,
		Body:         tpl.Body,
		EmailAccount: tpl.EmailAccount,
		IsActive:     tpl.IsActive,
	}, nil
}

// RegisterMessageService 注册消息 gRPC 服务到 gRPC Server。
func RegisterMessageService(server *grpc.Server, messageSvc *service.MessageService) {
	messagev1.RegisterMessageRPCServer(server, NewMessageRPCServer(messageSvc))
}