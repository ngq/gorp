// Application scenarios:
// - Expose root-package context helpers for container, validated body, request ID, and trace ID access.
// - Keep high-frequency support context accessors available from the top-level public API.
// - Re-export common support primitives without forcing direct dependency on framework/contract/support.
//
// 适用场景：
// - 暴露根包层的 container、validated body、request ID 和 trace ID context helper。
// - 让高频 support 上下文访问入口可以直接从顶层公共 API 使用。
// - 在不强迫业务直接依赖 framework/contract/support 的前提下重导出常用 support 原语。
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
