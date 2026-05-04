package support

import "context"

type containerContextKey struct{}

func NewContainerContext(ctx context.Context, c any) context.Context {
	return context.WithValue(ctx, containerContextKey{}, c)
}

func FromContainerContext(ctx context.Context) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	v := ctx.Value(containerContextKey{})
	return v, v != nil
}
