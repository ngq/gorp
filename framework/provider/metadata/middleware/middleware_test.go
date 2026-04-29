package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testMetadataPropagator struct{}

func (testMetadataPropagator) Inject(ctx context.Context, carrier contract.MetadataCarrier) {
	carrier.Set("x-md-user", "u1")
}

func (testMetadataPropagator) Extract(ctx context.Context, carrier contract.MetadataCarrier) context.Context {
	md := contract.NewMetadata()
	if v := carrier.Get("x-md-user"); v != "" {
		md.Set("x-md-user", v)
	}
	return contract.NewServerContext(ctx, md)
}

func TestMetadataHTTPMiddlewareExtractsIntoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	propagator := testMetadataPropagator{}
	r := gin.New()
	r.Use(MetadataMiddleware(propagator))

	var got string
	r.GET("/", func(c *gin.Context) {
		got = GetHeaderValue(c.Request.Context(), "x-md-user")
		c.Status(204)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("x-md-user", "u1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 204, w.Code)
	require.Equal(t, "u1", got)
}

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
