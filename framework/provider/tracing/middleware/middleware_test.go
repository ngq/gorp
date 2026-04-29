package middleware

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type testSpan struct {
	ctx    context.Context
	sc     contract.SpanContext
	status contract.SpanStatusCode
	msg    string
	err    error
	attrs  map[string]any
}

func (s *testSpan) End(options ...contract.SpanEndOption) {}
func (s *testSpan) AddEvent(name string, attributes map[string]interface{}) {}
func (s *testSpan) SetTag(key string, value interface{}) {}
func (s *testSpan) SpanContext() contract.SpanContext { return s.sc }
func (s *testSpan) SetStatus(code contract.SpanStatusCode, description string) { s.status, s.msg = code, description }
func (s *testSpan) SetAttributes(attrs map[string]any) {
	if s.attrs == nil { s.attrs = map[string]any{} }
	for k, v := range attrs { s.attrs[k] = v }
}
func (s *testSpan) SetError(err error) { s.err = err }
func (s *testSpan) IsRecording() bool { return true }
func (s *testSpan) Context() context.Context {
	if s.ctx == nil { return context.Background() }
	return s.ctx
}

type testTracer struct{}

func (testTracer) StartSpan(ctx context.Context, name string, opts ...contract.SpanOption) (context.Context, contract.Span) {
	cfg := &contract.SpanConfig{}
	for _, opt := range opts { opt(cfg) }
	span := &testSpan{ctx: ctx, sc: contract.SpanContext{TraceID: "trace-test"}, attrs: cfg.Attributes}
	return ctx, span
}
func (testTracer) SpanFromContext(ctx context.Context) contract.Span {
	return &testSpan{ctx: ctx, sc: contract.SpanContext{TraceID: "trace-test"}}
}
func (testTracer) Extract(ctx context.Context, carrier contract.TextMapCarrier) (context.Context, error) { return ctx, nil }
func (testTracer) Inject(ctx context.Context, carrier contract.TextMapCarrier) error { carrier.Set("traceparent", "tp"); return nil }

func TestTracingHTTPMiddlewareSetsTraceHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TracingMiddleware(testTracer{}, "svc"))
	r.GET("/ping", func(c *gin.Context) { c.Status(204) })

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	require.Equal(t, 204, w.Code)
	require.Equal(t, "trace-test", w.Header().Get("X-Trace-ID"))
}

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

func TestTracingGRPCServerInterceptorMarksErrorStatus(t *testing.T) {
	tracer := testTracer{}
	interceptor := UnaryServerInterceptor(tracer, "svc")
	expected := errors.New("boom")

	_, err := interceptor(context.Background(), nil, &ggrpc.UnaryServerInfo{FullMethod: "/demo.Service/Get"}, func(ctx context.Context, req any) (any, error) {
		return nil, expected
	})
	require.ErrorIs(t, err, expected)
}
