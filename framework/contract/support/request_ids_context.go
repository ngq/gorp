package support

import "context"

type requestIDContextKey struct{}
type traceIDContextKey struct{}

func NewRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

func FromRequestIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok
}

func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey{}, traceID)
}

func FromTraceIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	traceID, ok := ctx.Value(traceIDContextKey{}).(string)
	return traceID, ok
}
