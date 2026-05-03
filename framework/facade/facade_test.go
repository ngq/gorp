package facade

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/bootstrap"
	"github.com/ngq/gorp/framework/contract"
)

func TestHTTPOptionMapsToBootstrapOptions(t *testing.T) {
	cfg := runConfig{}
	HTTP(HTTPServiceOptions{DisableRedis: true, DisableGorm: true, DisableMetrics: true}).apply(&cfg)
	if !cfg.httpEnabled {
		t.Fatalf("expected http enabled")
	}
	if !cfg.httpOpts.DisableRedis || !cfg.httpOpts.DisableGorm || !cfg.httpOpts.DisableMetrics {
		t.Fatalf("expected HTTP options mapped to bootstrap options")
	}
}

func TestRunReturnsStableRunFailureError(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	cause := fmt.Errorf("run failed")
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		return cause
	}

	err := Run("demo")
	if !errors.Is(err, ErrHTTPServiceRunFailed) {
		t.Fatalf("expected ErrHTTPServiceRunFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped run cause, got %v", err)
	}
}

func TestStartReturnsStableRunFailureError(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	cause := fmt.Errorf("start failed")
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		return cause
	}

	err := Start("demo")
	if !errors.Is(err, ErrHTTPServiceRunFailed) {
		t.Fatalf("expected ErrHTTPServiceRunFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped start cause, got %v", err)
	}
}

func TestBuildReturnsStableBuildFailureError(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	cause := fmt.Errorf("build failed")
	newHTTPRuntimeFunc = func(serviceName string, opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		return nil, cause
	}

	_, err := Build("demo")
	if !errors.Is(err, ErrHTTPRuntimeBuildFailed) {
		t.Fatalf("expected ErrHTTPRuntimeBuildFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped build cause, got %v", err)
	}
}

func TestBuildHTTPRuntimeReturnsStableBuildFailureError(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	cause := fmt.Errorf("build runtime failed")
	newHTTPRuntimeFunc = func(serviceName string, opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		return nil, cause
	}

	_, err := BuildHTTPRuntime("demo")
	if !errors.Is(err, ErrHTTPRuntimeBuildFailed) {
		t.Fatalf("expected ErrHTTPRuntimeBuildFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped build runtime cause, got %v", err)
	}
}

func TestServiceIdentityContextHelpers(t *testing.T) {
	identity := &contract.ServiceIdentity{
		ServiceID:   "svc-1",
		ServiceName: "demo",
	}
	ctx := WithServiceIdentity(context.Background(), identity)

	got, ok := FromServiceIdentity(ctx)
	if !ok {
		t.Fatal("expected service identity from context")
	}
	if got.ServiceID != "svc-1" || got.ServiceName != "demo" {
		t.Fatalf("unexpected service identity: %#v", got)
	}
}

func TestWithProvidersAppendsNonNilProviders(t *testing.T) {
	cfg := runConfig{}
	p1 := bootstrap.FoundationProviders()[0]
	WithProviders(nil, p1).apply(&cfg)
	if len(cfg.httpOpts.ExtraProviders) != 1 {
		t.Fatalf("expected one provider, got %d", len(cfg.httpOpts.ExtraProviders))
	}
	if cfg.httpOpts.ExtraProviders[0].Name() != p1.Name() {
		t.Fatalf("expected provider %s", p1.Name())
	}
}

func TestWithProvidersDeduplicatesByName(t *testing.T) {
	cfg := runConfig{}
	p1 := bootstrap.FoundationProviders()[0]
	WithProviders(p1, p1).apply(&cfg)
	if len(cfg.httpOpts.ExtraProviders) != 1 {
		t.Fatalf("expected one provider after dedupe, got %d", len(cfg.httpOpts.ExtraProviders))
	}
	WithProviders(p1).apply(&cfg)
	if len(cfg.httpOpts.ExtraProviders) != 1 {
		t.Fatalf("expected still one provider after second append, got %d", len(cfg.httpOpts.ExtraProviders))
	}
}

func TestModuleAndWithModuleShareSameSemantics(t *testing.T) {
	cfg := runConfig{}
	p1 := bootstrap.FoundationProviders()[0]
	Module(p1).apply(&cfg)
	WithModule(p1).apply(&cfg)
	if len(cfg.httpOpts.ExtraProviders) != 1 {
		t.Fatalf("expected one provider after dedupe across module declarations, got %d", len(cfg.httpOpts.ExtraProviders))
	}
}

func TestModulesAppendsGroupedProviders(t *testing.T) {
	cfg := runConfig{}
	ps := bootstrap.FoundationProviders()
	if len(ps) < 2 {
		t.Fatalf("expected at least two foundation providers")
	}
	Modules([]ServiceProvider{ps[0]}, []ServiceProvider{ps[1]}).apply(&cfg)
	if len(cfg.httpOpts.ExtraProviders) != 2 {
		t.Fatalf("expected two grouped providers, got %d", len(cfg.httpOpts.ExtraProviders))
	}
}

func TestRunRequiresServiceName(t *testing.T) {
	err := Run("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestRunRejectsBlankServiceName(t *testing.T) {
	err := Run("   ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestStartRequiresServiceName(t *testing.T) {
	err := Start("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestOptionOrderWithoutHTTPThenHTTPIsRunnable(t *testing.T) {
	cfg, err := resolveRunConfig("demo", WithoutHTTP(), HTTP())
	if err != nil {
		t.Fatalf("expected runnable config, got %v", err)
	}
	if !cfg.httpEnabled {
		t.Fatalf("expected http enabled by last option")
	}
}

func TestOptionOrderHTTPThenWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := resolveRunConfig("demo", HTTP(), WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestRunWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Run("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestStartWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Start("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestRunContextHonorsPreCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, "demo")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestRunContextCanceledTakesPriorityOverServiceNameValidation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, "   ")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("startup cancellation should take priority over serviceName validation, got %v", err)
	}
}

func TestRunContextHonorsPreDeadlineExceededContext(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	err := RunContext(ctx, "demo")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}

func TestRunContextDeadlineExceededTakesPriorityOverServiceNameValidation(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	err := RunContext(ctx, "")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("startup deadline should take priority over serviceName validation, got %v", err)
	}
}

func TestRunContextNilContextUsesBackgroundAndBoots(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	err := RunContext(nil, "demo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatalf("expected bootHTTPService to be called")
	}
}

func TestRunContextCanceledSkipsBootHTTPService(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, "demo")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if called {
		t.Fatalf("bootHTTPService should not be called when startup context is canceled")
	}
}

func TestRunContextNilContextStillValidatesServiceName(t *testing.T) {
	err := RunContext(nil, " ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := RunContext(context.Background(), "demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestRunContextActiveBlankServiceNameReturnsServiceNameRequired(t *testing.T) {
	err := RunContext(context.Background(), " \t ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, "demo", WithoutHTTP())
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("startup cancellation should take priority over no-service-declared, got %v", err)
	}
}

func TestRunContextCanceledWithBlankServiceNameStillReturnsStartupCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, " ")
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("startup cancellation should take priority over serviceName validation, got %v", err)
	}
}

func TestRunBlankServiceNameSkipsBootHTTPService(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	err := Run("   ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
	if called {
		t.Fatalf("bootHTTPService should not be called when serviceName is blank")
	}
}

func TestRunWithoutHTTPSkipsBootHTTPService(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	err := Run("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
	if called {
		t.Fatalf("bootHTTPService should not be called when no service is declared")
	}
}

func TestRunContextUsesTrimmedServiceName(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	got := ""
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		got = serviceName
		return nil
	}

	err := RunContext(context.Background(), "  demo-service  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "demo-service" {
		t.Fatalf("expected trimmed serviceName demo-service, got %q", got)
	}
}

func TestBuildHTTPRuntimeRequiresServiceName(t *testing.T) {
	_, err := BuildHTTPRuntime("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestBuildRequiresServiceName(t *testing.T) {
	_, err := Build("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestBuildRejectsBlankServiceName(t *testing.T) {
	_, err := Build("  \t ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

func TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := BuildHTTPRuntime("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestBuildWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := Build("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

func TestBuildHTTPRuntimeUsesTrimmedServiceName(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	got := ""
	newHTTPRuntimeFunc = func(serviceName string, opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		got = serviceName
		return &bootstrap.HTTPServiceRuntime{}, nil
	}

	_, err := BuildHTTPRuntime("  build-service  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "build-service" {
		t.Fatalf("expected trimmed serviceName build-service, got %q", got)
	}
}

func TestBuildBlankServiceNameSkipsRuntimeBuilder(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	called := false
	newHTTPRuntimeFunc = func(serviceName string, opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		called = true
		return &bootstrap.HTTPServiceRuntime{}, nil
	}

	_, err := Build("  ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
	if called {
		t.Fatalf("newHTTPRuntimeFunc should not be called when serviceName is blank")
	}
}

func TestBuildWithoutHTTPSkipsRuntimeBuilder(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	called := false
	newHTTPRuntimeFunc = func(serviceName string, opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		called = true
		return &bootstrap.HTTPServiceRuntime{}, nil
	}

	_, err := Build("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
	if called {
		t.Fatalf("newHTTPRuntimeFunc should not be called when no service is declared")
	}
}

func TestWithSetupComposesInOptionOrder(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "a")
		return nil
	}).apply(&cfg)
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "b")
		return nil
	}).apply(&cfg)
	if cfg.setup == nil {
		t.Fatalf("expected setup composed")
	}
	if err := cfg.setup(&HTTPRuntime{}); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	if len(seq) != 2 || seq[0] != "a" || seq[1] != "b" {
		t.Fatalf("expected setup execution order [a b], got %v", seq)
	}
}

func TestWithHTTPRoutesComposesWithSetup(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "setup")
		return nil
	}).apply(&cfg)
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
		seq = append(seq, "r1")
		return nil
	}).apply(&cfg)
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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

func TestWithSetupShortCircuitOnError(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	cause := fmt.Errorf("boom")
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "first")
		return cause
	}).apply(&cfg)
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
		seq = append(seq, "routes")
		return nil
	}).apply(&cfg)
	err := cfg.setup(&HTTPRuntime{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrSetupFailed) {
		t.Fatalf("expected ErrSetupFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped setup cause, got %v", err)
	}
	if len(seq) != 1 || seq[0] != "first" {
		t.Fatalf("expected short-circuit after first error, got %v", seq)
	}
}

func TestWithSetupNilDoesNotBreakExistingSetupChain(t *testing.T) {
	cfg := runConfig{}
	called := 0
	WithSetup(func(rt *HTTPRuntime) error {
		called++
		return nil
	}).apply(&cfg)
	WithSetup(nil).apply(&cfg)
	if cfg.setup == nil {
		t.Fatalf("expected setup chain to stay available")
	}
	if err := cfg.setup(&HTTPRuntime{}); err != nil {
		t.Fatalf("unexpected setup error: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected previous setup called once, got %d", called)
	}
}

func TestWithMigrateWrapsErrorWithStableContract(t *testing.T) {
	cfg := runConfig{}
	cause := fmt.Errorf("migrate boom")
	WithMigrate(func(rt *HTTPRuntime) error {
		return cause
	}).apply(&cfg)
	if cfg.migrate == nil {
		t.Fatalf("expected migrate hook assigned")
	}
	err := cfg.migrate(&HTTPRuntime{})
	if !errors.Is(err, ErrMigrateFailed) {
		t.Fatalf("expected ErrMigrateFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped migrate cause, got %v", err)
	}
}

func TestWithMigrateNilDoesNotOverrideExistingMigrate(t *testing.T) {
	cfg := runConfig{}
	called := 0
	WithMigrate(func(rt *HTTPRuntime) error {
		called++
		return nil
	}).apply(&cfg)
	WithMigrate(nil).apply(&cfg)
	if cfg.migrate == nil {
		t.Fatalf("expected migrate hook to stay available")
	}
	if err := cfg.migrate(&HTTPRuntime{}); err != nil {
		t.Fatalf("unexpected migrate error: %v", err)
	}
	if called != 1 {
		t.Fatalf("expected previous migrate called once, got %d", called)
	}
}

func TestWithMigrateLastWins(t *testing.T) {
	cfg := runConfig{}
	called := make([]string, 0, 2)
	WithMigrate(func(rt *HTTPRuntime) error {
		called = append(called, "first")
		return nil
	}).apply(&cfg)
	WithMigrate(func(rt *HTTPRuntime) error {
		called = append(called, "second")
		return nil
	}).apply(&cfg)
	if err := cfg.migrate(&HTTPRuntime{}); err != nil {
		t.Fatalf("unexpected migrate error: %v", err)
	}
	if len(called) != 1 || called[0] != "second" {
		t.Fatalf("expected only last migrate called, got %v", called)
	}
}

func TestHTTPOptionLastWins(t *testing.T) {
	cfg, err := resolveRunConfig(
		"demo",
		HTTP(HTTPServiceOptions{DisableRedis: true, DisableGorm: true, DisableMetrics: true}),
		HTTP(HTTPServiceOptions{DisableRedis: false, DisableGorm: true, DisableMetrics: false}),
	)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if cfg.httpOpts.DisableRedis {
		t.Fatalf("expected DisableRedis from last HTTP option to be false")
	}
	if !cfg.httpOpts.DisableGorm {
		t.Fatalf("expected DisableGorm from last HTTP option to be true")
	}
	if cfg.httpOpts.DisableMetrics {
		t.Fatalf("expected DisableMetrics from last HTTP option to be false")
	}
}

func TestWithHTTPRoutesWrapsRouteRegistrationError(t *testing.T) {
	cfg := runConfig{}
	cause := fmt.Errorf("register routes failed")
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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

func testHTTPRuntime() *HTTPRuntime {
	return &HTTPRuntime{
		Router:    testRouter{},
		Container: testContainer{},
	}
}

type testRouter struct{}

func (testRouter) Use(middleware ...contract.HTTPMiddleware) {}
func (testRouter) Group(prefix string, middleware ...contract.HTTPMiddleware) contract.HTTPRouter {
	return testRouter{}
}
func (testRouter) Handle(method, path string, handler contract.HTTPHandler)        {}
func (testRouter) HandleFunc(method, path string, handlerFunc contract.HTTPHandler) {}
func (testRouter) GET(path string, handler contract.HTTPHandler)                   {}
func (testRouter) POST(path string, handler contract.HTTPHandler)                  {}
func (testRouter) PUT(path string, handler contract.HTTPHandler)                   {}
func (testRouter) DELETE(path string, handler contract.HTTPHandler)                {}
func (testRouter) Mount(path string, handler http.Handler)                         {}

type testContainer struct{}

func (testContainer) Bind(key string, factory contract.Factory, singleton bool) {}
func (testContainer) IsBind(key string) bool                                  { return false }
func (testContainer) Make(key string) (any, error)                            { return nil, nil }
func (testContainer) MustMake(key string) any                                 { return nil }
func (testContainer) RegisterProvider(p contract.ServiceProvider) error       { return nil }
func (testContainer) RegisterProviders(providers ...contract.ServiceProvider) error {
	return nil
}

func TestWithHTTPRoutesReturnsRuntimeUnavailableWhenRuntimeIsNil(t *testing.T) {
	cfg := runConfig{}
	called := false
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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
	WithHTTPRoutes(func(router contract.HTTPRouter, container contract.Container) error {
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
