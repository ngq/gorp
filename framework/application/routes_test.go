// Package application_test provides unit tests for application route registration and composition.
//
// 适用场景：
// - 验证 HTTP Route Option 的组合、注册和执行顺序。
package application

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// TestWithHTTPRoutesComposesWithSetup verifies that WithHTTPRoutes composes
// with WithSetup and executes them in the correct order.
//
// TestWithHTTPRoutesComposesWithSetup 验证 WithHTTPRoutes 与 WithSetup 组合
// 并按正确顺序执行。
func TestWithHTTPRoutesComposesWithSetup(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "setup")
		return nil
	}).apply(&cfg)
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		seq = append(seq, "routes")
		return nil
	}).apply(&cfg)
	if cfg.setup == nil {
		t.Fatalf("expected setup composed")
	}
	if err := cfg.setup(testHTTPRuntime()); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	if len(seq) != 2 || seq[0] != "setup" || seq[1] != "routes" {
		t.Fatalf("expected execution order [setup routes], got %v", seq)
	}
}

// TestWithHTTPRoutesNilRegistrarIsNoOp verifies that passing a nil registrar
// to WithHTTPRoutes results in a no-op setup function.
//
// TestWithHTTPRoutesNilRegistrarIsNoOp 验证向 WithHTTPRoutes 传递 nil 注册器
// 会导致无操作设置函数。
func TestWithHTTPRoutesNilRegistrarIsNoOp(t *testing.T) {
	cfg := runConfig{}
	WithHTTPRoutes(nil).apply(&cfg)
	if cfg.setup == nil {
		t.Fatalf("expected setup to be assigned as no-op")
	}
	if err := cfg.setup(&HTTPRuntime{}); err != nil {
		t.Fatalf("expected nil error for no-op routes, got %v", err)
	}
}

// TestWithHTTPRoutesExecuteInOptionOrder verifies that multiple WithHTTPRoutes
// options execute their registrars in the order they were applied.
//
// TestWithHTTPRoutesExecuteInOptionOrder 验证多个 WithHTTPRoutes 选项
// 按应用顺序执行其注册器。
func TestWithHTTPRoutesExecuteInOptionOrder(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		seq = append(seq, "r1")
		return nil
	}).apply(&cfg)
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		seq = append(seq, "r2")
		return nil
	}).apply(&cfg)
	if err := cfg.setup(testHTTPRuntime()); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	if len(seq) != 2 || seq[0] != "r1" || seq[1] != "r2" {
		t.Fatalf("expected route execution order [r1 r2], got %v", seq)
	}
}

// TestWithHTTPRoutesWrapsRouteRegistrationError verifies that when a route
// registrar returns an error, it is wrapped in ErrHTTPRouteRegistrationFailed.
//
// TestWithHTTPRoutesWrapsRouteRegistrationError 验证当路由注册器返回错误时，
// 该错误会被包装在 ErrHTTPRouteRegistrationFailed 中。
func TestWithHTTPRoutesWrapsRouteRegistrationError(t *testing.T) {
	cfg := runConfig{}
	cause := fmt.Errorf("register routes failed")
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		return cause
	}).apply(&cfg)
	err := cfg.setup(testHTTPRuntime())
	if !errors.Is(err, ErrHTTPRouteRegistrationFailed) {
		t.Fatalf("expected ErrHTTPRouteRegistrationFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped route cause, got %v", err)
	}
}

// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenRuntimeIsNil verifies that
// setup returns ErrHTTPRuntimeUnavailable when the runtime is nil.
//
// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenRuntimeIsNil 验证当 runtime
// 为 nil 时，设置返回 ErrHTTPRuntimeUnavailable。
func TestWithHTTPRoutesReturnsRuntimeUnavailableWhenRuntimeIsNil(t *testing.T) {
	cfg := runConfig{}
	called := false
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		called = true
		return nil
	}).apply(&cfg)
	err := cfg.setup(nil)
	if !errors.Is(err, ErrHTTPRuntimeUnavailable) {
		t.Fatalf("expected ErrHTTPRuntimeUnavailable, got %v", err)
	}
	if called {
		t.Fatalf("expected registrar not called when runtime is nil")
	}
}

// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenEngineIsNil verifies that
// setup returns ErrHTTPRuntimeUnavailable when the router engine is nil.
//
// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenEngineIsNil 验证当路由引擎
// 为 nil 时，设置返回 ErrHTTPRuntimeUnavailable。
func TestWithHTTPRoutesReturnsRuntimeUnavailableWhenEngineIsNil(t *testing.T) {
	cfg := runConfig{}
	called := false
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		called = true
		return nil
	}).apply(&cfg)
	err := cfg.setup(&HTTPRuntime{})
	if !errors.Is(err, ErrHTTPRuntimeUnavailable) {
		t.Fatalf("expected ErrHTTPRuntimeUnavailable, got %v", err)
	}
	if called {
		t.Fatalf("expected registrar not called when router is nil")
	}
}

// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenContainerIsNil verifies that
// setup returns ErrHTTPRuntimeUnavailable when the container is nil.
//
// TestWithHTTPRoutesReturnsRuntimeUnavailableWhenContainerIsNil 验证当容器
// 为 nil 时，设置返回 ErrHTTPRuntimeUnavailable。
func TestWithHTTPRoutesReturnsRuntimeUnavailableWhenContainerIsNil(t *testing.T) {
	cfg := runConfig{}
	called := false
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
		called = true
		return nil
	}).apply(&cfg)
	err := cfg.setup(&HTTPRuntime{Router: testRouter{}})
	if !errors.Is(err, ErrHTTPRuntimeUnavailable) {
		t.Fatalf("expected ErrHTTPRuntimeUnavailable, got %v", err)
	}
	if called {
		t.Fatalf("expected registrar not called when container is nil")
	}
}

func testHTTPRuntime() *HTTPRuntime {
	return &HTTPRuntime{
		Router:    testRouter{},
		Container: testContainer{},
	}
}

type testRouter struct{}

func (testRouter) Use(middleware ...transportcontract.HTTPMiddleware) {}
func (testRouter) Group(prefix string, middleware ...transportcontract.HTTPMiddleware) transportcontract.HTTPRouter {
	return testRouter{}
}
func (testRouter) Handle(method, path string, handler transportcontract.HTTPHandler)         {}
func (testRouter) HandleFunc(method, path string, handlerFunc transportcontract.HTTPHandler) {}
func (testRouter) GET(path string, handler transportcontract.HTTPHandler)                    {}
func (testRouter) POST(path string, handler transportcontract.HTTPHandler)                   {}
func (testRouter) PUT(path string, handler transportcontract.HTTPHandler)                    {}
func (testRouter) DELETE(path string, handler transportcontract.HTTPHandler)                 {}
func (testRouter) Mount(path string, handler http.Handler)                                   {}

type testContainer struct{}

func (testContainer) Bind(key string, factory runtimecontract.Factory, singleton bool)            {}
func (testContainer) NamedBind(name, key string, factory runtimecontract.Factory, singleton bool) {}
func (testContainer) IsBind(key string) bool                                                      { return false }
func (testContainer) IsBindNamed(name, key string) bool                                           { return false }
func (testContainer) Make(key string) (any, error)                                                { return nil, nil }
func (testContainer) MakeNamed(name, key string) (any, error)                                     { return nil, nil }
func (testContainer) MustMake(key string) any                                                     { return nil }
func (testContainer) MustMakeNamed(name, key string) any                                          { return nil }
func (testContainer) RegisterCloser(key string, closer io.Closer)                                 {}
func (testContainer) Destroy() error                                                              { return nil }
func (testContainer) RegisterProvider(p runtimecontract.ServiceProvider) error                    { return nil }
func (testContainer) RegisterProviders(providers ...runtimecontract.ServiceProvider) error {
	return nil
}
func (testContainer) RegisteredProviders() []runtimecontract.ProviderInfo { return nil }
func (testContainer) DebugPrint() string                                  { return "" }
func (testContainer) ProviderDAG() runtimecontract.ProviderDAG             { return runtimecontract.ProviderDAG{} }
