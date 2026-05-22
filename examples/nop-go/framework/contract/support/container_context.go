// Application scenarios:
// - Attach the runtime container onto request-scoped or task-scoped contexts.
// - Give middleware, handlers, and background jobs a lightweight way to access container-bound capabilities.
// - Keep container context keys private and collision-free.
//
// 适用场景：
// - 将运行时容器挂到请求级或任务级 context 上。
// - 为 middleware、handler 和后台任务提供轻量级容器访问路径。
// - 保持容器相关 context key 私有且避免冲突。
package support

import "context"

type containerContextKey struct{}

// NewContainerContext stores a container-like object into the context.
//
// NewContainerContext 将容器对象写入 context。
func NewContainerContext(ctx context.Context, c any) context.Context {
	return context.WithValue(ctx, containerContextKey{}, c)
}

// FromContainerContext reads a container-like object from the context.
//
// FromContainerContext 从 context 中读取容器对象。
func FromContainerContext(ctx context.Context) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(containerContextKey{})
	return v, v != nil
}
