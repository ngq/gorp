package support

import "context"

type validatedBodyContextKey struct{}

func NewValidatedBodyContext(ctx context.Context, body any) context.Context {
	return context.WithValue(ctx, validatedBodyContextKey{}, body)
}

func FromValidatedBodyContext(ctx context.Context) (any, bool) {
	if ctx == nil {
		return nil, false
	}
	body := ctx.Value(validatedBodyContextKey{})
	return body, body != nil
}
