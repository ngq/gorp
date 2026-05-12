// Package application_test provides unit tests for application route registration and composition.
//
// 适用场景：
// - 验证 HTTP Route Option 的组合、注册和执行顺序。
package application

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

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

func (testContainer) Bind(key string, factory runtimecontract.Factory, singleton bool) {}
func (testContainer) IsBind(key string) bool                                           { return false }
func (testContainer) Make(key string) (any, error)                                     { return nil, nil }
func (testContainer) MustMake(key string) any                                          { return nil }
func (testContainer) RegisterProvider(p runtimecontract.ServiceProvider) error         { return nil }
func (testContainer) RegisterProviders(providers ...runtimecontract.ServiceProvider) error {
	return nil
}
