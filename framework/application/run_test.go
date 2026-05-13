// Package application_test provides unit tests for application run lifecycle and error handling.
//
// 适用场景：
// - 验证 Application Run 的启动、停止和错误处理行为。
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

// TestRunReturnsStableRunFailureError verifies that Run wraps boot failures in ErrHTTPServiceRunFailed.
//
// TestRunReturnsStableRunFailureError 验证 Run 将启动失败包装为 ErrHTTPServiceRunFailed。
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

// TestStartReturnsStableRunFailureError verifies that Start wraps boot failures in ErrHTTPServiceRunFailed.
//
// TestStartReturnsStableRunFailureError 验证 Start 将启动失败包装为 ErrHTTPServiceRunFailed。
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

// TestBuildReturnsStableBuildFailureError verifies that Build wraps runtime build failures in ErrHTTPRuntimeBuildFailed.
//
// TestBuildReturnsStableBuildFailureError 验证 Build 将运行时构建失败包装为 ErrHTTPRuntimeBuildFailed。
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

// TestBuildHTTPRuntimeReturnsStableBuildFailureError verifies that BuildHTTPRuntime wraps build failures in ErrHTTPRuntimeBuildFailed.
//
// TestBuildHTTPRuntimeReturnsStableBuildFailureError 验证 BuildHTTPRuntime 将构建失败包装为 ErrHTTPRuntimeBuildFailed。
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

// TestRunRequiresServiceName verifies that Run returns ErrServiceNameRequired when service name is empty.
//
// TestRunRequiresServiceName 验证 Run 在服务名为空时返回 ErrServiceNameRequired。
func TestRunRequiresServiceName(t *testing.T) {
	err := Run("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestRunRejectsBlankServiceName verifies that Run returns ErrServiceNameRequired when service name is blank.
//
// TestRunRejectsBlankServiceName 验证 Run 在服务名为空白时返回 ErrServiceNameRequired。
func TestRunRejectsBlankServiceName(t *testing.T) {
	err := Run("   ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestStartRequiresServiceName verifies that Start returns ErrServiceNameRequired when service name is empty.
//
// TestStartRequiresServiceName 验证 Start 在服务名为空时返回 ErrServiceNameRequired。
func TestStartRequiresServiceName(t *testing.T) {
	err := Start("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestRunWithoutHTTPReturnsNoServiceDeclared verifies that Run returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestRunWithoutHTTPReturnsNoServiceDeclared 验证 Run 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestRunWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Run("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestStartWithoutHTTPReturnsNoServiceDeclared verifies that Start returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestStartWithoutHTTPReturnsNoServiceDeclared 验证 Start 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestStartWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Start("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestRunContextHonorsPreCanceledContext verifies that RunContext returns ErrStartupCanceled when context is already canceled.
//
// TestRunContextHonorsPreCanceledContext 验证 RunContext 在 context 已取消时返回 ErrStartupCanceled。
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

// TestRunContextCanceledTakesPriorityOverServiceNameValidation verifies that startup cancellation is checked before service name validation.
//
// TestRunContextCanceledTakesPriorityOverServiceNameValidation 验证启动取消检查优先于服务名验证。
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

// TestRunContextHonorsPreDeadlineExceededContext verifies that RunContext returns ErrStartupCanceled when context deadline is exceeded.
//
// TestRunContextHonorsPreDeadlineExceededContext 验证 RunContext 在 context 超时时返回 ErrStartupCanceled。
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

// TestRunContextDeadlineExceededTakesPriorityOverServiceNameValidation verifies that deadline exceeded is checked before service name validation.
//
// TestRunContextDeadlineExceededTakesPriorityOverServiceNameValidation 验证截止时间检查优先于服务名验证。
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

// TestRunContextNilContextUsesBackgroundAndBoots verifies that RunContext with nil context uses Background and boots the service.
//
// TestRunContextNilContextUsesBackgroundAndBoots 验证 RunContext 在 nil context 时使用 Background 并启动服务。
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

// TestRunContextCanceledSkipsBootHTTPService verifies that bootHTTPService is not called when startup context is canceled.
//
// TestRunContextCanceledSkipsBootHTTPService 验证启动 context 取消时跳过 bootHTTPService 调用。
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

// TestRunContextNilContextStillValidatesServiceName verifies that nil context still triggers service name validation.
//
// TestRunContextNilContextStillValidatesServiceName 验证 nil context 仍然执行服务名验证。
func TestRunContextNilContextStillValidatesServiceName(t *testing.T) {
	err := RunContext(nil, " ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared verifies that RunContext returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared 验证 RunContext 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := RunContext(context.Background(), "demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestRunContextActiveBlankServiceNameReturnsServiceNameRequired verifies that blank service name returns ErrServiceNameRequired.
//
// TestRunContextActiveBlankServiceNameReturnsServiceNameRequired 验证空白服务名返回 ErrServiceNameRequired。
func TestRunContextActiveBlankServiceNameReturnsServiceNameRequired(t *testing.T) {
	err := RunContext(context.Background(), " \t ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled verifies that cancellation takes priority over no-service-declared error.
//
// TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled 验证取消错误优先于无服务声明错误。
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

// TestRunContextCanceledWithBlankServiceNameStillReturnsStartupCanceled verifies that cancellation takes priority over service name validation.
//
// TestRunContextCanceledWithBlankServiceNameStillReturnsStartupCanceled 验证取消错误优先于服务名验证。
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

// TestRunBlankServiceNameSkipsBootHTTPService verifies that bootHTTPService is not called when service name is blank.
//
// TestRunBlankServiceNameSkipsBootHTTPService 验证服务名为空白时跳过 bootHTTPService 调用。
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

// TestRunWithoutHTTPSkipsBootHTTPService verifies that bootHTTPService is not called when no HTTP service is declared.
//
// TestRunWithoutHTTPSkipsBootHTTPService 验证未声明 HTTP 服务时跳过 bootHTTPService 调用。
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

// TestRunContextUsesTrimmedServiceName verifies that RunContext trims whitespace from service name before passing to bootHTTPService.
//
// TestRunContextUsesTrimmedServiceName 验证 RunContext 将服务名去空白后再传递给 bootHTTPService。
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

// TestRunContextLeavesGovernanceOverrideEmptyByDefault verifies that governance mode override is empty by default.
//
// TestRunContextLeavesGovernanceOverrideEmptyByDefault 验证治理模式覆盖默认为空。
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

// TestRunContextForwardsMicroserviceGovernanceMode verifies that WithMicroserviceMode option sets governance mode correctly.
//
// TestRunContextForwardsMicroserviceGovernanceMode 验证 WithMicroserviceMode 选项正确设置治理模式。
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

// TestBuildHTTPRuntimeRequiresServiceName verifies that BuildHTTPRuntime returns ErrServiceNameRequired when service name is empty.
//
// TestBuildHTTPRuntimeRequiresServiceName 验证 BuildHTTPRuntime 在服务名为空时返回 ErrServiceNameRequired。
func TestBuildHTTPRuntimeRequiresServiceName(t *testing.T) {
	_, err := BuildHTTPRuntime("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestBuildRequiresServiceName verifies that Build returns ErrServiceNameRequired when service name is empty.
//
// TestBuildRequiresServiceName 验证 Build 在服务名为空时返回 ErrServiceNameRequired。
func TestBuildRequiresServiceName(t *testing.T) {
	_, err := Build("")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestBuildRejectsBlankServiceName verifies that Build returns ErrServiceNameRequired when service name is blank.
//
// TestBuildRejectsBlankServiceName 验证 Build 在服务名为空白时返回 ErrServiceNameRequired。
func TestBuildRejectsBlankServiceName(t *testing.T) {
	_, err := Build("  \t ")
	if !errors.Is(err, ErrServiceNameRequired) {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}
}

// TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared verifies that BuildHTTPRuntime returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared 验证 BuildHTTPRuntime 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := BuildHTTPRuntime("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestBuildWithoutHTTPReturnsNoServiceDeclared verifies that Build returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestBuildWithoutHTTPReturnsNoServiceDeclared 验证 Build 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestBuildWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := Build("demo", WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestBuildHTTPRuntimeUsesTrimmedServiceName verifies that BuildHTTPRuntime trims whitespace from service name.
//
// TestBuildHTTPRuntimeUsesTrimmedServiceName 验证 BuildHTTPRuntime 对服务名进行去空白处理。
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

// TestBuildBlankServiceNameSkipsRuntimeBuilder verifies that newHTTPRuntimeFunc is not called when service name is blank.
//
// TestBuildBlankServiceNameSkipsRuntimeBuilder 验证服务名为空白时跳过 newHTTPRuntimeFunc 调用。
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

// TestBuildWithoutHTTPSkipsRuntimeBuilder verifies that newHTTPRuntimeFunc is not called when no HTTP service is declared.
//
// TestBuildWithoutHTTPSkipsRuntimeBuilder 验证未声明 HTTP 服务时跳过 newHTTPRuntimeFunc 调用。
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
