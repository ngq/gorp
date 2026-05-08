package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/bootstrap"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

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

func TestRunContextLeavesGovernanceOverrideEmptyByDefault(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()

	gotMode := "<unset>"
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		gotMode = opts.GovernanceMode
		return nil
	}

	err := RunContext(context.Background(), "demo", HTTP())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMode != "" {
		t.Fatalf("expected default governance override to be empty, got %q", gotMode)
	}
}

func TestRunContextForwardsMicroserviceGovernanceMode(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()

	gotMode := ""
	bootHTTPService = func(serviceName string, opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		gotMode = opts.GovernanceMode
		return nil
	}

	err := RunContext(context.Background(), "demo", HTTP(), WithMicroserviceMode())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMode != string(resiliencecontract.GovernanceModeMicroservice) {
		t.Fatalf("expected governance mode microservice, got %q", gotMode)
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
