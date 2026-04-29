package grpc

import (
	"context"
	"testing"

	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"github.com/stretchr/testify/require"
)

func TestUnaryServerInterceptorExtractsTraceAndRequestID(t *testing.T) {
	interceptor := UnaryServerInterceptor()
	ctx := metadata.NewIncomingContext(context.Background(), metadata.New(map[string]string{
		TraceIDKey:   "trace-1",
		RequestIDKey: "req-1",
	}))

	_, err := interceptor(ctx, nil, &ggrpc.UnaryServerInfo{FullMethod: "/demo.Service/Get"}, func(ctx context.Context, req interface{}) (interface{}, error) {
		require.Equal(t, "trace-1", GetTraceID(ctx))
		require.Equal(t, "req-1", GetRequestID(ctx))
		return "ok", nil
	})
	require.NoError(t, err)
}

func TestUnaryClientInterceptorInjectsOutgoingMetadata(t *testing.T) {
	interceptor := UnaryClientInterceptor()
	ctx := context.WithValue(context.Background(), traceIDContextKey, "trace-2")
	ctx = context.WithValue(ctx, requestIDContextKey, "req-2")

	err := interceptor(ctx, "/demo.Service/Get", nil, nil, nil, func(ctx context.Context, method string, req, reply interface{}, cc *ggrpc.ClientConn, opts ...ggrpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		require.Equal(t, []string{"trace-2"}, md.Get(TraceIDKey))
		require.Equal(t, []string{"req-2"}, md.Get(RequestIDKey))
		return nil
	})
	require.NoError(t, err)
}
