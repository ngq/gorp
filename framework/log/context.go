package log

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

type contextLoggerKey struct{}

// WithContext 把请求级 logger 写入 context。
//
// 中文说明：
// - 供 framework transport / middleware 层注入请求级 logger；
// - 业务层读取统一走 Ctx(ctx)。
func WithContext(ctx context.Context, l contract.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = Default()
	}
	return context.WithValue(ctx, contextLoggerKey{}, l)
}

// FromContext 从 context 中读取请求级 logger。
func FromContext(ctx context.Context) (contract.Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	l, ok := ctx.Value(contextLoggerKey{}).(contract.Logger)
	return l, ok && l != nil
}

// Ctx 返回当前 context 关联的 logger。
//
// 中文说明：
// - 优先返回请求级 logger；
// - 取不到时回退到 Default()，保证业务层总能拿到可用 logger。
func Ctx(ctx context.Context) contract.Logger {
	if l, ok := FromContext(ctx); ok {
		return l
	}
	return Default()
}

// WithContextFields 基于 context 关联的 logger 追加字段。
//
// 中文说明：
// - 用于减少业务层频繁书写 `Ctx(ctx).With(...)`；
// - 无请求级 logger 时自动回退到 Default()。
func WithContextFields(ctx context.Context, fields ...contract.Field) contract.Logger {
	return Ctx(ctx).With(fields...)
}
