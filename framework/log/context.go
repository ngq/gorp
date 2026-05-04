package log

import (
	"context"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
)

type contextLoggerKey struct{}

func WithContext(ctx context.Context, l observabilitycontract.Logger) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if l == nil {
		l = Default()
	}
	return context.WithValue(ctx, contextLoggerKey{}, l)
}

func FromContext(ctx context.Context) (observabilitycontract.Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	l, ok := ctx.Value(contextLoggerKey{}).(observabilitycontract.Logger)
	return l, ok && l != nil
}

func Ctx(ctx context.Context) observabilitycontract.Logger {
	if l, ok := FromContext(ctx); ok {
		return l
	}
	return Default()
}

func WithContextFields(ctx context.Context, fields ...observabilitycontract.Field) observabilitycontract.Logger {
	return Ctx(ctx).With(fields...)
}
