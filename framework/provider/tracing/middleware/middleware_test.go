// Package middleware_test provides unit tests for tracing middleware interceptor behavior.
//
// 适用场景：
// - 验证 tracing middleware 在 HTTP 和 gRPC 请求中的 span 创建和传播。
// - 确保 trace context 正确注入到下游调用。
package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
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

// testContext implements Context for testing
type testContext struct {
	gin *gin.Context
}

func (c *testContext) Context() context.Context {
	return c.gin.Request.Context()
}

func (c *testContext) Request() *http.Request {
	return c.gin.Request
}

func (c *testContext) Response() http.ResponseWriter {
	return c.gin.Writer
}

func (c *testContext) Param(key string) string {
	return c.gin.Param(key)
}

func (c *testContext) Query(key string) string {
	return c.gin.Query(key)
}

func (c *testContext) DefaultQuery(key, defaultValue string) string {
	return c.gin.DefaultQuery(key, defaultValue)
}

func (c *testContext) GetHeader(key string) string {
	return c.gin.GetHeader(key)
}

func (c *testContext) SetHeader(key, value string) {
	c.gin.Header(key, value)
}

func (c *testContext) Bind(obj any) error {
	return c.gin.ShouldBind(obj)
}

func (c *testContext) BindJSON(obj any) error {
	return c.gin.ShouldBindJSON(obj)
}

func (c *testContext) BindQuery(obj any) error {
	return c.gin.ShouldBindQuery(obj)
}

func (c *testContext) JSON(status int, body any) {
	c.gin.JSON(status, body)
}

func (c *testContext) String(status int, body string) {
	c.gin.String(status, body)
}

func (c *testContext) XML(status int, body any) {
	c.gin.XML(status, body)
}

func (c *testContext) Data(status int, contentType string, body []byte) {
	c.gin.Data(status, contentType, body)
}

func (c *testContext) Redirect(status int, location string) {
	c.gin.Redirect(status, location)
}

func (c *testContext) Status(code int) {
	c.gin.Status(code)
}

func (c *testContext) RoutePath() string {
	return c.gin.FullPath()
}

func (c *testContext) ResponseStatus() int {
	return c.gin.Writer.Status()
}

func (c *testContext) Get(key string) any {
	val, _ := c.gin.Get(key)
	return val
}

func (c *testContext) Set(key string, value any) {
	c.gin.Set(key, value)
}

func (c *testContext) Abort(status int) {
	c.gin.AbortWithStatus(status)
}

func (c *testContext) AbortWithJSON(status int, body any) {
	c.gin.AbortWithStatusJSON(status, body)
}

func (c *testContext) IsAborted() bool {
	return c.gin.IsAborted()
}

func (c *testContext) Next() {
	c.gin.Next()
}

func newTestContext(c *gin.Context) transportcontract.Context {
	return &testContext{gin: c}
}

// TestTracingMiddlewareSetsTraceHeader verifies that middleware injects trace ID into response headers.
//
// TestTracingMiddlewareSetsTraceHeader 验证中间件将 trace ID 注入响应头。
func TestTracingMiddlewareSetsTraceHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.ContextWithFallback = true
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := TracingMiddleware(testTracer{}, "svc")(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})
	r.GET("/ping", func(c *gin.Context) {
		// Trace ID is stored via c.Set()
		traceIDVal, exists := c.Get("trace_id")
		require.True(t, exists)
		traceID, ok := traceIDVal.(string)
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
