// Package rpc 提供 trade-service 调用其他微服务的 gRPC 客户端封装。
// 包含通用的连接管理能力，以及各服务客户端的工厂函数。
package rpc

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial 根据 gRPC 地址创建客户端连接。
// 使用不安全凭据（内部服务间调用无需 TLS），后续可按需替换为 TLS。
func Dial(addr string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	allOpts := append(defaultOpts, opts...)
	conn, err := grpc.NewClient(addr, allOpts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC dial %s: %w", addr, err)
	}
	return conn, nil
}