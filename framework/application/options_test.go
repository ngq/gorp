package application

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ngq/gorp/framework/bootstrap"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

func TestHTTPOptionMapsToBootstrapOptions(t *testing.T) {
	cfg := runConfig{}
	HTTP(HTTPServiceOptions{DisableRedis: true, DisableGorm: true, DisableMetrics: true, GovernanceMode: resiliencecontract.GovernanceModeMicroservice}).apply(&cfg)
	if !cfg.httpEnabled {
		t.Fatalf("expected http enabled")
	}
	if !cfg.httpOpts.DisableRedis || !cfg.httpOpts.DisableGorm || !cfg.httpOpts.DisableMetrics {
		t.Fatalf("expected HTTP options mapped to bootstrap options")
	}
	if cfg.httpOpts.GovernanceMode != "microservice" {
		t.Fatalf("expected governance mode propagated, got %q", cfg.httpOpts.GovernanceMode)
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

func TestWithSetupShortCircuitOnError(t *testing.T) {
	cfg := runConfig{}
	seq := make([]string, 0, 2)
	cause := fmt.Errorf("boom")
	WithSetup(func(rt *HTTPRuntime) error {
		seq = append(seq, "first")
		return cause
	}).apply(&cfg)
	WithHTTPRoutes(func(router transportcontract.HTTPRouter, container runtimecontract.Container) error {
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

func TestWithGovernanceModeOverridesStartupMode(t *testing.T) {
	cfg, err := resolveRunConfig(
		"demo",
		HTTP(),
		WithMicroserviceMode(),
	)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if cfg.httpOpts.GovernanceMode != "microservice" {
		t.Fatalf("expected governance mode microservice, got %q", cfg.httpOpts.GovernanceMode)
	}

	WithMonolithMode().apply(&cfg)
	if cfg.httpOpts.GovernanceMode != "monolith" {
		t.Fatalf("expected governance mode monolith after override, got %q", cfg.httpOpts.GovernanceMode)
	}
}

func TestHTTPOptionGovernanceModeLastWinsAcrossMultipleHTTPDeclarations(t *testing.T) {
	cfg, err := resolveRunConfig(
		"demo",
		HTTP(HTTPServiceOptions{GovernanceMode: resiliencecontract.GovernanceModeMonolith}),
		HTTP(HTTPServiceOptions{GovernanceMode: resiliencecontract.GovernanceModeMicroservice}),
	)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if cfg.httpOpts.GovernanceMode != "microservice" {
		t.Fatalf("expected last HTTP governance mode to win, got %q", cfg.httpOpts.GovernanceMode)
	}
}

func TestExplicitGovernanceModeOptionOverridesHTTPGovernanceMode(t *testing.T) {
	cfg, err := resolveRunConfig(
		"demo",
		HTTP(HTTPServiceOptions{GovernanceMode: resiliencecontract.GovernanceModeMonolith}),
		WithGovernanceMode(resiliencecontract.GovernanceModeMicroservice),
	)
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if cfg.httpOpts.GovernanceMode != "microservice" {
		t.Fatalf("expected explicit governance mode option to win, got %q", cfg.httpOpts.GovernanceMode)
	}
}
