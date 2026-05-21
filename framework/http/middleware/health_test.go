// Package middleware_test provides tests for HTTP health check endpoints.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockHealthChecker 用于测试
type mockHealthChecker struct {
	report *observabilitycontract.HealthReport
	err    error
}

func (m *mockHealthChecker) Check(ctx context.Context) (*observabilitycontract.HealthReport, error) {
	return m.report, m.err
}

func (m *mockHealthChecker) CheckComponent(ctx context.Context, name string) (*observabilitycontract.HealthCheckResult, error) {
	return nil, nil
}

func (m *mockHealthChecker) AddChecker(name string, checker observabilitycontract.ComponentChecker) {}

func (m *mockHealthChecker) AddDependency(name string, dep observabilitycontract.DependencyChecker) {}

// TestHealthCheckHandler_Healthy 测试健康状态返回 200
func TestHealthCheckHandler_Healthy(t *testing.T) {
	checker := &mockHealthChecker{
		report: &observabilitycontract.HealthReport{
			Service:      "test-service",
			Version:      "1.0.0",
			Status:       observabilitycontract.HealthStatusHealthy,
			Checks:       map[string]observabilitycontract.HealthCheckResult{},
			Dependencies: map[string]observabilitycontract.DependencyHealth{},
		},
	}

	router := NewTestEngine()
	router.GET("/healthz", func(c *gin.Context) {
		HealthCheckHandler(checker)(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var resp observabilitycontract.HealthReport
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Service != "test-service" {
		t.Errorf("Service = %v, want test-service", resp.Service)
	}
}

// TestHealthCheckHandler_Unhealthy 测试不健康状态返回 503
func TestHealthCheckHandler_Unhealthy(t *testing.T) {
	checker := &mockHealthChecker{
		report: &observabilitycontract.HealthReport{
			Service:      "test-service",
			Version:      "1.0.0",
			Status:       observabilitycontract.HealthStatusUnhealthy,
			Checks:       map[string]observabilitycontract.HealthCheckResult{},
			Dependencies: map[string]observabilitycontract.DependencyHealth{},
		},
	}

	router := NewTestEngine()
	router.GET("/healthz", func(c *gin.Context) {
		HealthCheckHandler(checker)(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusServiceUnavailable)
	}
}

// TestHealthCheckHandler_NilChecker 测试 nil checker 返回 503
func TestHealthCheckHandler_NilChecker(t *testing.T) {
	router := NewTestEngine()
	router.GET("/healthz", func(c *gin.Context) {
		HealthCheckHandler(nil)(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusServiceUnavailable)
	}
}

// TestReadinessHandler_Healthy 测试就绪检查 healthy 返回 200
func TestReadinessHandler_Healthy(t *testing.T) {
	checker := &mockHealthChecker{
		report: &observabilitycontract.HealthReport{
			Service:      "test-service",
			Status:       observabilitycontract.HealthStatusHealthy,
			Checks:       map[string]observabilitycontract.HealthCheckResult{},
			Dependencies: map[string]observabilitycontract.DependencyHealth{},
		},
	}

	router := NewTestEngine()
	router.GET("/readyz", func(c *gin.Context) {
		ReadinessHandler(checker)(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["ready"] != true {
		t.Errorf("ready = %v, want true", resp["ready"])
	}
}

// TestReadinessHandler_Degraded 测试就绪检查 degraded 返回 503
func TestReadinessHandler_Degraded(t *testing.T) {
	checker := &mockHealthChecker{
		report: &observabilitycontract.HealthReport{
			Service:      "test-service",
			Status:       observabilitycontract.HealthStatusDegraded,
			Checks:       map[string]observabilitycontract.HealthCheckResult{},
			Dependencies: map[string]observabilitycontract.DependencyHealth{},
		},
	}

	router := NewTestEngine()
	router.GET("/readyz", func(c *gin.Context) {
		ReadinessHandler(checker)(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusServiceUnavailable)
	}
}

// TestLivenessHandler 测试存活检查返回 200
func TestLivenessHandler(t *testing.T) {
	router := NewTestEngine()
	router.GET("/livez", func(c *gin.Context) {
		LivenessHandler()(newContext(c))
	})

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp["alive"] != true {
		t.Errorf("alive = %v, want true", resp["alive"])
	}
}

// TestRegisterHealthEndpointsSimple 测试简单健康端点注册
func TestRegisterHealthEndpointsSimple(t *testing.T) {
	// 创建 mock router
	mockRouter := &mockRouter{}

	RegisterHealthEndpointsSimple(mockRouter, "test-service", "1.0.0")

	// 验证端点已注册
	if len(mockRouter.routes) != 3 {
		t.Errorf("Registered routes count = %v, want 3", len(mockRouter.routes))
	}

	expectedRoutes := []string{"/healthz", "/readyz", "/livez"}
	for _, expected := range expectedRoutes {
		if _, ok := mockRouter.routes[expected]; !ok {
			t.Errorf("Route %v not registered", expected)
		}
	}
}

// mockRouter 用于测试
type mockRouter struct {
	routes map[string]transportcontract.Handler
}

func (m *mockRouter) GET(path string, handler transportcontract.Handler) {
	if m.routes == nil {
		m.routes = make(map[string]transportcontract.Handler)
	}
	m.routes[path] = handler
}

func (m *mockRouter) POST(path string, handler transportcontract.Handler)    {}
func (m *mockRouter) PUT(path string, handler transportcontract.Handler)     {}
func (m *mockRouter) DELETE(path string, handler transportcontract.Handler)  {}
func (m *mockRouter) PATCH(path string, handler transportcontract.Handler)   {}
func (m *mockRouter) OPTIONS(path string, handler transportcontract.Handler) {}
func (m *mockRouter) HEAD(path string, handler transportcontract.Handler)    {}
func (m *mockRouter) ANY(path string, handler transportcontract.Handler)     {}
func (m *mockRouter) Use(middleware ...transportcontract.Middleware)         {}
func (m *mockRouter) Group(prefix string, middleware ...transportcontract.Middleware) transportcontract.Router {
	return m
}
func (m *mockRouter) Mount(prefix string, handler http.Handler) {}
func (m *mockRouter) Handle(method, path string, handler transportcontract.Handler) {
	m.GET(path, handler)
}
func (m *mockRouter) HandleFunc(method, path string, handler transportcontract.Handler) {
	m.GET(path, handler)
}
func (m *mockRouter) Static(relativePath, root string)                 {}
func (m *mockRouter) StaticFile(relativePath, filepath string)         {}
func (m *mockRouter) StaticFS(relativePath string, fs http.FileSystem) {}
