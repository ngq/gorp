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
