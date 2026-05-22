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
)

// TestRunReturnsStableRunFailureError verifies that Run wraps boot failures in ErrHTTPServiceRunFailed.
//
// TestRunReturnsStableRunFailureError 验证 Run 将启动失败包装为 ErrHTTPServiceRunFailed。
func TestRunReturnsStableRunFailureError(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	cause := fmt.Errorf("run failed")
	bootHTTPService = func(opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		return cause
	}

	err := Run()
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
	bootHTTPService = func(opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		return cause
	}

	err := Start()
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
	newHTTPRuntimeFunc = func(opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		return nil, cause
	}

	_, err := Build()
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
	newHTTPRuntimeFunc = func(opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		return nil, cause
	}

	_, err := BuildHTTPRuntime()
	if !errors.Is(err, ErrHTTPRuntimeBuildFailed) {
		t.Fatalf("expected ErrHTTPRuntimeBuildFailed, got %v", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("expected wrapped build runtime cause, got %v", err)
	}
}

// TestRunWithoutHTTPReturnsNoServiceDeclared verifies that Run returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestRunWithoutHTTPReturnsNoServiceDeclared 验证 Run 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestRunWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Run(WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestStartWithoutHTTPReturnsNoServiceDeclared verifies that Start returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestStartWithoutHTTPReturnsNoServiceDeclared 验证 Start 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestStartWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := Start(WithoutHTTP())
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
	err := RunContext(ctx)
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

// TestRunContextHonorsPreDeadlineExceededContext verifies that RunContext returns ErrStartupCanceled when context deadline is exceeded.
//
// TestRunContextHonorsPreDeadlineExceededContext 验证 RunContext 在 context 超时时返回 ErrStartupCanceled。
func TestRunContextHonorsPreDeadlineExceededContext(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	err := RunContext(ctx)
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline exceeded, got %v", err)
	}
}

// TestRunContextNilContextUsesBackgroundAndBoots verifies that RunContext with nil context uses Background and boots the service.
//
// TestRunContextNilContextUsesBackgroundAndBoots 验证 RunContext 在 nil context 时使用 Background 并启动服务。
func TestRunContextNilContextUsesBackgroundAndBoots(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	err := RunContext(nil)
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
	bootHTTPService = func(opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx)
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if called {
		t.Fatalf("bootHTTPService should not be called when startup context is canceled")
	}
}

// TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared verifies that RunContext returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared 验证 RunContext 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestRunContextActiveWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	err := RunContext(context.Background(), WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled verifies that cancellation takes priority over no-service-declared error.
//
// TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled 验证取消错误优先于无服务声明错误。
func TestRunContextCanceledWithWithoutHTTPStillReturnsStartupCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := RunContext(ctx, WithoutHTTP())
	if !errors.Is(err, ErrStartupCanceled) {
		t.Fatalf("expected ErrStartupCanceled, got %v", err)
	}
	if errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("startup cancellation should take priority over no-service-declared, got %v", err)
	}
}

// TestRunWithoutHTTPSkipsBootHTTPService verifies that bootHTTPService is not called when no HTTP service is declared.
//
// TestRunWithoutHTTPSkipsBootHTTPService 验证未声明 HTTP 服务时跳过 bootHTTPService 调用。
func TestRunWithoutHTTPSkipsBootHTTPService(t *testing.T) {
	origin := bootHTTPService
	defer func() { bootHTTPService = origin }()
	called := false
	bootHTTPService = func(opts bootstrap.HTTPServiceOptions, migrate func(*bootstrap.HTTPServiceRuntime) error, setup func(*bootstrap.HTTPServiceRuntime) error) error {
		called = true
		return nil
	}

	err := Run(WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
	if called {
		t.Fatalf("bootHTTPService should not be called when no service is declared")
	}
}

// TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared verifies that BuildHTTPRuntime returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared 验证 BuildHTTPRuntime 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestBuildHTTPRuntimeWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := BuildHTTPRuntime(WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestBuildWithoutHTTPReturnsNoServiceDeclared verifies that Build returns ErrNoServiceDeclared when HTTP is disabled.
//
// TestBuildWithoutHTTPReturnsNoServiceDeclared 验证 Build 在禁用 HTTP 时返回 ErrNoServiceDeclared。
func TestBuildWithoutHTTPReturnsNoServiceDeclared(t *testing.T) {
	_, err := Build(WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
}

// TestBuildWithoutHTTPSkipsRuntimeBuilder verifies that newHTTPRuntimeFunc is not called when no HTTP service is declared.
//
// TestBuildWithoutHTTPSkipsRuntimeBuilder 验证未声明 HTTP 服务时跳过 newHTTPRuntimeFunc 调用。
func TestBuildWithoutHTTPSkipsRuntimeBuilder(t *testing.T) {
	origin := newHTTPRuntimeFunc
	defer func() { newHTTPRuntimeFunc = origin }()
	called := false
	newHTTPRuntimeFunc = func(opts bootstrap.HTTPServiceOptions) (*bootstrap.HTTPServiceRuntime, error) {
		called = true
		return &bootstrap.HTTPServiceRuntime{}, nil
	}

	_, err := Build(WithoutHTTP())
	if !errors.Is(err, ErrNoServiceDeclared) {
		t.Fatalf("expected ErrNoServiceDeclared, got %v", err)
	}
	if called {
		t.Fatalf("newHTTPRuntimeFunc should not be called when no service is declared")
	}
}
