package contract

import "context"

var requestContainerKey = &contextKey{name: "request-container"}

// NewContainerContext stores the runtime container in request context.
func NewContainerContext(ctx context.Context, c Container) context.Context {
	return context.WithValue(ctx, requestContainerKey, c)
}

// FromContainerContext retrieves the runtime container from request context.
func FromContainerContext(ctx context.Context) (Container, bool) {
	container, ok := ctx.Value(requestContainerKey).(Container)
	return container, ok
}
