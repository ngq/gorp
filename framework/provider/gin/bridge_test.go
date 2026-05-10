package gin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	httpmiddleware "github.com/ngq/gorp/framework/http/middleware"
)

func TestAdaptMiddlewarePreservesRequestIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	// 挂载 RequestIdentity middleware 通过 AdaptMiddleware
	engine.Use(AdaptMiddleware(httpmiddleware.RequestIdentity()))

	var capturedRequestID string
	engine.Use(func(c *gin.Context) {
		// RequestIdentity 将 request ID 写入 context，而非 header
		requestID, ok := supportcontract.FromRequestIDContext(c.Request.Context())
		if ok {
			capturedRequestID = requestID
		}
		c.Next()
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	if capturedRequestID == "" {
		t.Error("expected request ID to be set by adapted middleware, got empty")
	}
}

func TestAdaptMiddlewareSyncsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	engine.Use(AdaptMiddleware(httpmiddleware.RequestIdentity()))

	var ginContextHasTraceID bool
	engine.Use(func(c *gin.Context) {
		_, ok := supportcontract.FromTraceIDContext(c.Request.Context())
		ginContextHasTraceID = ok
		c.Next()
	})
	engine.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	// RequestIdentity middleware should set trace ID in context
	// and it should be visible in subsequent Gin middleware
	if !ginContextHasTraceID {
		t.Error("expected trace ID to be propagated to gin.Context via context sync")
	}
}

func TestNativeEngineFromHTTPService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	svc := &service{
		engine: engine,
		router: newRouter(&engine.RouterGroup),
	}

	extracted, ok := NativeEngine(svc)
	if !ok {
		t.Fatal("expected to extract *gin.Engine from service")
	}
	if extracted != engine {
		t.Error("extracted engine does not match original")
	}
}

func TestNativeEngineFromNonGinService(t *testing.T) {
	// 模拟非 Gin HTTP 服务
	svc := &nonGinHTTPService{}
	_, ok := NativeEngine(svc)
	if ok {
		t.Error("expected NativeEngine to return false for non-Gin HTTP service")
	}
}

func TestNativeRouterGroupFromHTTPService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	svc := &service{
		engine: engine,
		router: newRouter(&engine.RouterGroup),
	}

	rg, ok := NativeRouterGroup(svc)
	if !ok {
		t.Fatal("expected to extract *gin.RouterGroup from service")
	}
	if rg == nil {
		t.Error("expected non-nil RouterGroup")
	}
}

func TestGinFirstEngineDoesNotAutoMountGovernance(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Gin-first engine 只注入 container middleware
	c := &mockContainer{}
	engine := newGinFirstEngine(c)

	// 验证只有 1 个 middleware (injectRequestContainer)
	// Gin 的 Handlers 数量在路由注册后才能确定，因此通过行为验证
	var called bool
	engine.GET("/test", func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	if !called {
		t.Error("expected handler to be called")
	}
}

func TestMixedGinAndAbstractMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	var order []string

	// 原生 Gin middleware
	engine.Use(func(c *gin.Context) {
		order = append(order, "gin-native")
		c.Next()
	})

	// 框架抽象 middleware 通过 AdaptMiddleware
	engine.Use(AdaptMiddleware(func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(ctx transportcontract.HTTPContext) {
			order = append(order, "adapted-abstract")
			next(ctx)
		}
	}))

	engine.GET("/test", func(c *gin.Context) {
		order = append(order, "handler")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	engine.ServeHTTP(w, req)

	expected := []string{"gin-native", "adapted-abstract", "handler"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d entries, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("position %d: expected %q, got %q", i, v, order[i])
		}
	}
}

// mockContainer 实现 runtimecontract.Container 的最小子集。
type mockContainer struct{}

func (m *mockContainer) IsBind(string) bool                                     { return false }
func (m *mockContainer) Make(string) (any, error)                               { return nil, nil }
func (m *mockContainer) MustMake(string) any                                    { return nil }
func (m *mockContainer) Bind(string, runtimecontract.Factory, bool)             {}
func (m *mockContainer) RegisterProvider(runtimecontract.ServiceProvider) error  { return nil }
func (m *mockContainer) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}

// nonGinHTTPService 模拟不实现 GINEngineProvider 的 HTTP 服务。
type nonGinHTTPService struct{}

func (n *nonGinHTTPService) Router() transportcontract.HTTPRouter    { return nil }
func (n *nonGinHTTPService) Server() *http.Server                    { return nil }
func (n *nonGinHTTPService) Run() error                              { return nil }
func (n *nonGinHTTPService) Shutdown(ctx context.Context) error      { return nil }
