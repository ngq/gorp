package contract

import "context"

type validatedBodyContextKey struct{}

// NewValidatedBodyContext stores a validated request body in request context.
func NewValidatedBodyContext(ctx context.Context, body any) context.Context {
	return context.WithValue(ctx, validatedBodyContextKey{}, body)
}

// FromValidatedBodyContext retrieves a validated request body from request context.
func FromValidatedBodyContext(ctx context.Context) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	body := ctx.Value(validatedBodyContextKey{})
	if body == nil {
		return nil, false
	}
	return body, true
}
