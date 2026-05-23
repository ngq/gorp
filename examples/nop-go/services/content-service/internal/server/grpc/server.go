// Package grpc 实现 content-service 的 gRPC 服务端。
// ContentRPCServer 实现 contentv1.UnimplementedContentRPCServer 接口，
// 将 gRPC 请求委托给已有的 ContentService 处理。
package grpc

import (
	"context"

	contentv1 "nop-go/services/content-service/api/content/v1"
	"nop-go/services/content-service/internal/service"

	"google.golang.org/grpc"
)

// ContentRPCServer 内容服务 gRPC 服务端实现。
type ContentRPCServer struct {
	contentv1.UnimplementedContentRPCServer
	contentSvc *service.ContentService
}

// NewContentRPCServer 创建内容服务 gRPC 服务端。
func NewContentRPCServer(contentSvc *service.ContentService) *ContentRPCServer {
	return &ContentRPCServer{contentSvc: contentSvc}
}

// GetBlog 根据 ID 获取博客内容 —— 供 admin 等服务获取关联内容。
func (s *ContentRPCServer) GetBlog(ctx context.Context, req *contentv1.GetBlogReq) (*contentv1.GetBlogResp, error) {
	blog, err := s.contentSvc.GetBlog(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &contentv1.GetBlogResp{
		Id:        blog.ID,
		Title:     blog.Title,
		Content:   blog.Content,
		Author:    blog.Author,
		Status:    blog.Status,
		Tags:      blog.Tags,
		CreatedAt: blog.CreatedAt,
		UpdatedAt: blog.UpdatedAt,
	}, nil
}

// RegisterContentService 注册内容 gRPC 服务到 gRPC Server。
func RegisterContentService(server *grpc.Server, contentSvc *service.ContentService) {
	contentv1.RegisterContentRPCServer(server, NewContentRPCServer(contentSvc))
}