// Package gorp provides the root-package application startup surface for gorp framework.
// This file exposes context helpers for container, validated body, request ID, trace ID.
// Keeps high-frequency support context accessors available from top-level API.
//
// Gorp 包提供 gorp 框架的根包层应用启动入口。
// 本文件暴露根包层的 container、validated body、request ID、trace ID context helper。
// 让高频 support 上下文访问入口可以直接从顶层公共 API 使用。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/contract/support"
)

func NewContainerContext(ctx context.Context, c Container) context.Context {
	return support.NewContainerContext(ctx, c)
}

func FromContainerContext(ctx context.Context) (Container, bool) {
	v, ok := support.FromContainerContext(ctx)
	if !ok {
		return nil, false
	}
	c, ok := v.(runtime.Container)
	return c, ok
}

func FromValidatedBodyContext(ctx context.Context) (any, bool) {
	return support.FromValidatedBodyContext(ctx)
}

func FromRequestIDContext(ctx context.Context) (string, bool) {
	return support.FromRequestIDContext(ctx)
}

func FromTraceIDContext(ctx context.Context) (string, bool) {
	return support.FromTraceIDContext(ctx)
}
