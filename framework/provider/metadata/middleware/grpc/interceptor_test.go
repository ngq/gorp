package grpc

import (
	"context"
	"testing"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testMetadataPropagator struct{}

func (testMetadataPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
	carrier.Set("x-md-user", "u1")
}

func (testMetadataPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	md := transportcontract.NewMetadata()
	if v := carrier.Get("x-md-user"); v != "" {
		md.Set("x-md-user", v)
	}
	return transportcontract.NewServerContext(ctx, md)
}

// TestMetadataGRPCClientInterceptorPreservesOutgoingMetadata 验证 gRPC 客户端拦截器保留已有 metadata。
func TestMetadataGRPCClientInterceptorPreservesOutgoingMetadata(t *testing.T) {
	propagator := testMetadataPropagator{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"x-existing": "keep"}))

	err := UnaryClientInterceptor(propagator)(ctx, "/demo.Service/Get", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *ggrpc.ClientConn, opts ...ggrpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		require.Equal(t, []string{"keep"}, md.Get("x-existing"))
		require.Equal(t, []string{"u1"}, md.Get("x-md-user"))
		return nil
	})
	require.NoError(t, err)
}
