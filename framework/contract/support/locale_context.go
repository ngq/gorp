// Application scenarios:
// - Propagate request-scoped locale information through context.
// - Let validation, response shaping, and business handlers read the effective locale consistently.
// - Keep locale-related context keys private and collision-free.
//
// 适用场景：
// - 在 context 中透传请求级 locale 信息。
// - 让校验、响应整形和业务 handler 统一读取当前 locale。
// - 保持 locale 相关 context key 私有且避免冲突。
package support

import "context"

type localeContextKey struct{}

// NewLocaleContext stores locale information in context.
//
// NewLocaleContext 将 locale 信息写入 context。
func NewLocaleContext(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, localeContextKey{}, locale)
}

// FromLocaleContext reads locale information from context.
//
// FromLocaleContext 从 context 中读取 locale 信息。
func FromLocaleContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	locale, ok := ctx.Value(localeContextKey{}).(string)
	return locale, ok && locale != ""
}
