// Package grpc 提供 grpc-free 契约的断言辅助函数。
package grpc

import (
	"fmt"

	"google.golang.org/grpc"
)

// MustGRPCServer 把契约层返回的 any 断言为 *grpc.Server。
// 契约层为保持 grpc-free 用 any 暴露，proto-first 应用注册服务时用它转换。
//
// MustGRPCServer 将 any 断言为 *grpc.Server。
func MustGRPCServer(v any) *grpc.Server {
	if s, ok := v.(*grpc.Server); ok {
		return s
	}
	panic(fmt.Sprintf("expected *grpc.Server, got %T", v))
}

// MustGRPCConn 把契约层返回的 any 断言为 *grpc.ClientConn。
//
// MustGRPCConn 将 any 断言为 *grpc.ClientConn。
func MustGRPCConn(v any) *grpc.ClientConn {
	if c, ok := v.(*grpc.ClientConn); ok {
		return c
	}
	panic(fmt.Sprintf("expected *grpc.ClientConn, got %T", v))
}
