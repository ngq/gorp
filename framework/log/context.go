// Application scenarios:
// - Attach request-scoped loggers to context for middleware and handler flows.
// - Let downstream code fetch either the context logger or the default fallback consistently.
// - Support deriving per-request field-enriched loggers without mutating the global default logger.
//
// 适用场景：
// - 在 context 中挂载请求级 logger，供 middleware 和 handler 流程使用。
// - 让下游代码一致地获取 context logger 或默认回退 logger。
// - 支持在不修改全局默认 logger 的前提下派生带请求字段的 logger。
package log

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

type contextLoggerKey struct{}

// WithContext stores one logger into the context.
//
// WithContext 将一个 logger 写入 context。
func WithContext(ctx context.Context, l observabilitycontract.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = Default()
	}
	return context.WithValue(ctx, contextLoggerKey{}, l)
}

// FromContext reads the logger stored in the context.
//
// FromContext 从 context 中读取 logger。
func FromContext(ctx context.Context) (observabilitycontract.Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	l, ok := ctx.Value(contextLoggerKey{}).(observabilitycontract.Logger)
	return l, ok && l != nil
}

// Ctx returns the context logger or falls back to the default logger.
//
// Ctx 返回 context logger；若不存在则回退到默认 logger。
func Ctx(ctx context.Context) observabilitycontract.Logger {
	if l, ok := FromContext(ctx); ok {
		return l
	}
	return Default()
}

// WithContextFields derives a logger from the context logger and appends fields.
//
// WithContextFields 基于 context logger 派生一个追加了字段的新 logger。
func WithContextFields(ctx context.Context, fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return Ctx(ctx).With(fields...)
}
