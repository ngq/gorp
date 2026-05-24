// Package middleware_test provides unit tests for metadata propagation middleware.
//
// 适用场景：
// - 验证 metadata propagator 在 HTTP 和 gRPC 之间的上下文传播行为。
// - 确保 carrier 实现和 middleware 拦截逻辑正确。
package middleware

import (
	"context"
	"mime/multipart"
	"net/http"
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

func (c *testContext) Get(key string) (any, bool) {
	return c.gin.Get(key)
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

func (c *testContext) DefaultIntQuery(key string, defaultValue int) int {
	return defaultValue
}

func (c *testContext) Int64Param(key string) (int64, error) {
	return 0, nil
}

func (c *testContext) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	return nil, nil, http.ErrNoCookie
}

func (c *testContext) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return nil
}

func newTestContext(c *gin.Context) transportcontract.Context {
	return &testContext{gin: c}
}

// TestMetadataMiddlewareExtractsIntoContext 验证 metadata 中间件正确提取 header 到 context。
//
// 中文说明：
// - MetadataMiddleware 从 HTTP header 中提取 metadata 并写入 context。
// - 下游 handler 可通过 GetHeaderValue 获取提取的 metadata。
func TestMetadataMiddlewareExtractsIntoContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	propagator := testMetadataPropagator{}
	r := gin.New()
	r.ContextWithFallback = true
	r.Use(func(c *gin.Context) {
		httpCtx := newTestContext(c)
		wrapped := MetadataMiddleware(propagator)(func(inner transportcontract.Context) {
			c.Next()
		})
		if wrapped != nil {
			wrapped(httpCtx)
		}
	})

	var got string
	r.GET("/", func(c *gin.Context) {
		// Metadata is stored via c.Set()
		if mdVal, exists := c.Get("metadata"); exists {
			if md, ok := mdVal.(transportcontract.Metadata); ok {
				got = md.Get("x-md-user")
			}
		}
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
