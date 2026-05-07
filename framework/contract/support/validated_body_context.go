// Application scenarios:
// - Store validated request payloads in context for downstream reuse.
// - Avoid repeated bind/validate work across middleware and handlers.
// - Keep validated-body context keys private and collision-free.
//
// 适用场景：
// - 在 context 中存储已经完成校验的请求体，供下游复用。
// - 避免 middleware 和 handler 之间重复做 bind/validate。
// - 保持 validated-body context key 私有且避免冲突。
package support

import "context"

type validatedBodyContextKey struct{}

// NewValidatedBodyContext stores a validated body in context.
//
// NewValidatedBodyContext 将已校验请求体写入 context。
func NewValidatedBodyContext(ctx context.Context, body any) context.Context {
	return context.WithValue(ctx, validatedBodyContextKey{}, body)
}

// FromValidatedBodyContext reads a validated body from context.
//
// FromValidatedBodyContext 从 context 中读取已校验请求体。
func FromValidatedBodyContext(ctx context.Context) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	body := ctx.Value(validatedBodyContextKey{})
	return body, body != nil
}
