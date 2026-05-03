package contract

import "context"

type requestIDContextKey struct{}
type traceIDContextKey struct{}

// NewRequestIDContext stores request id in context.
func NewRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDContextKey{}, requestID)
}

// FromRequestIDContext retrieves request id from context.
func FromRequestIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok && requestID != ""
}

// NewTraceIDContext stores trace id in context.
func NewTraceIDContext(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey{}, traceID)
}

// FromTraceIDContext retrieves trace id from context.
func FromTraceIDContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	traceID, ok := ctx.Value(traceIDContextKey{}).(string)
	return traceID, ok && traceID != ""
}
