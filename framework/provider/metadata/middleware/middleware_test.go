// Package middleware_test provides unit tests for metadata propagation middleware.
//
// 适用场景：
// - 验证 metadata propagator 在 HTTP 和 gRPC 之间的上下文传播行为。
// - 确保 carrier 实现和 middleware 拦截逻辑正确。
package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

// TestMetadataHTTPMiddlewareExtractsIntoContext 验证 HTTP metadata 中间件正确提取 header 到 context。
//
// 中文说明：
// - MetadataMiddleware 从 HTTP header 中提取 metadata 并写入 context。
// - 下游 handler 可通过 GetHeaderValue 获取提取的 metadata。
func TestMetadataHTTPMiddlewareExtractsIntoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	propagator := testMetadataPropagator{}
	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		wrapped := MetadataMiddleware(propagator)(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})

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

// TestMetadataGRPCClientInterceptorPreservesOutgoingMetadata 验证 gRPC 客户端拦截器保留已有 metadata。
//
// 中文说明：
// - UnaryClientInterceptor 在传播 metadata 时保留已有的 outgoing context。
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
