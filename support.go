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

// ResolveResourcePath safely searches for a file path across standard runtime locations:
// 1. Direct path (absolute or relative to current working directory)
// 2. Relative to APP_BASE_PATH
// 3. Extra search directories
// 4. Running executable directory
// 5. Parent directories of current working directory
//
// ResolveResourcePath 安全解析文件路径，跨执行目录、APP_BASE_PATH、候选目录及父级目录智能探测。
func ResolveResourcePath(filePath string, extraSearchDirs ...string) string {
	return support.ResolveResourcePath(filePath, extraSearchDirs...)
}
