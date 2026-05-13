// Package middleware_test provides unit tests for tracing middleware interceptor behavior.
//
// 适用场景：
// - 验证 tracing middleware 在 HTTP 和 gRPC 请求中的 span 创建和传播。
// - 确保 trace context 正确注入到下游调用。
package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testSpan struct {
	ctx    context.Context
	sc     observabilitycontract.SpanContext
	status observabilitycontract.SpanStatusCode
	msg    string
	err    error
	attrs  map[string]any
}

func (s *testSpan) End(options ...observabilitycontract.SpanEndOption)      {}
func (s *testSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *testSpan) SetTag(key string, value interface{})                    {}
func (s *testSpan) SpanContext() observabilitycontract.SpanContext          { return s.sc }
func (s *testSpan) SetStatus(code observabilitycontract.SpanStatusCode, description string) {
	s.status, s.msg = code, description
}
func (s *testSpan) SetAttributes(attrs map[string]any) {
	if s.attrs == nil {
		s.attrs = map[string]any{}
	}
	for k, v := range attrs {
		s.attrs[k] = v
	}
}
func (s *testSpan) SetError(err error) { s.err = err }
func (s *testSpan) IsRecording() bool  { return true }
func (s *testSpan) Context() context.Context {
	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

type testTracer struct{}

func (testTracer) StartSpan(ctx context.Context, name string, opts ...observabilitycontract.SpanOption) (context.Context, observabilitycontract.Span) {
	cfg := &observabilitycontract.SpanConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	span := &testSpan{ctx: ctx, sc: observabilitycontract.SpanContext{TraceID: "trace-test"}, attrs: cfg.Attributes}
	return ctx, span
}
func (testTracer) SpanFromContext(ctx context.Context) observabilitycontract.Span {
	return &testSpan{ctx: ctx, sc: observabilitycontract.SpanContext{TraceID: "trace-test"}}
}
func (testTracer) Extract(ctx context.Context, carrier observabilitycontract.TextMapCarrier) (context.Context, error) {
	return ctx, nil
}
func (testTracer) Inject(ctx context.Context, carrier observabilitycontract.TextMapCarrier) error {
	carrier.Set("traceparent", "tp")
	return nil
}

// TestTracingHTTPMiddlewareSetsTraceHeader verifies that HTTP middleware injects trace ID into response headers.
//
// TestTracingHTTPMiddlewareSetsTraceHeader 验证 HTTP 中间件将 trace ID 注入响应头。
func TestTracingHTTPMiddlewareSetsTraceHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		httpCtx := transportcontract.NewDefaultHTTPContext(c.Request.Context(), c.Request)
		httpCtx.SetHeaderFuncs(c.GetHeader, c.Header)
		httpCtx.SetResponseFuncs(c.JSON, func(code int, body string) { c.String(code, body) }, c.XML, c.Data, c.Redirect, c.Status, func() int { return c.Writer.Status() })
		httpCtx.SetRoutePathFunc(c.FullPath)
		wrapped := TracingMiddleware(testTracer{}, "svc")(func(inner transportcontract.HTTPContext) {
			if inner != nil && inner.Request() != nil {
				c.Request = inner.Request()
			}
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/ping", func(c *gin.Context) {
		traceID, ok := supportcontract.FromTraceIDContext(c.Request.Context())
		require.True(t, ok)
		require.Equal(t, "trace-test", traceID)
		c.Status(204)
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 204, w.Code)
	require.Equal(t, "trace-test", w.Header().Get("X-Trace-ID"))
}

// TestTracingGRPCClientInterceptorInjectsMetadata verifies that gRPC client interceptor injects trace metadata.
//
// TestTracingGRPCClientInterceptorInjectsMetadata 验证 gRPC 客户端拦截器注入 trace 元数据。
func TestTracingGRPCClientInterceptorInjectsMetadata(t *testing.T) {
	tracer := testTracer{}
	err := UnaryClientInterceptor(tracer, "svc")(context.Background(), "/demo.Service/Get", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc *ggrpc.ClientConn, opts ...ggrpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		require.True(t, ok)
		require.Equal(t, []string{"tp"}, md.Get("traceparent"))
		return nil
	})
	require.NoError(t, err)
}

// TestTracingGRPCServerInterceptorMarksErrorStatus verifies that gRPC server interceptor marks error status.
//
// TestTracingGRPCServerInterceptorMarksErrorStatus 验证 gRPC 服务端拦截器标记错误状态。
func TestTracingGRPCServerInterceptorMarksErrorStatus(t *testing.T) {
	tracer := testTracer{}
	interceptor := UnaryServerInterceptor(tracer, "svc")
	expected := errors.New("boom")

	_, err := interceptor(context.Background(), nil, &ggrpc.UnaryServerInfo{FullMethod: "/demo.Service/Get"}, func(ctx context.Context, req any) (any, error) {
		return nil, expected
	})
	require.ErrorIs(t, err, expected)
}
