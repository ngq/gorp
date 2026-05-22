// Application scenarios:
// - Propagate request ID and trace ID through request-scoped context.
// - Give middleware, logging, tracing, and handlers one shared access path to correlation identifiers.
// - Keep identifier context keys private to avoid collisions across packages.
//
// 适用场景：
// - 在请求级 context 中透传 request ID 和 trace ID。
// - 为 middleware、日志、tracing 和 handler 提供统一的关联标识读取路径。
// - 保持标识相关 context key 私有，避免跨包冲突。
package support

import "context"

type requestIDContextKey struct{}
type traceIDContextKey struct{}

// NewRequestIDContext stores request ID in context.
//
// NewRequestIDContext 将 request ID 写入 context。
func NewRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

// FromRequestIDContext reads request ID from context.
//
// FromRequestIDContext 从 context 中读取 request ID。
func FromRequestIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok
}

// NewTraceIDContext stores trace ID in context.
//
// NewTraceIDContext 将 trace ID 写入 context。
func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey{}, traceID)
}

// FromTraceIDContext reads trace ID from context.
//
// FromTraceIDContext 从 context 中读取 trace ID。
func FromTraceIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	traceID, ok := ctx.Value(traceIDContextKey{}).(string)
	return traceID, ok
}
