// Package host_test provides unit tests for the host provider.
//
// 适用场景：
// - 验证 Host provider 的注册、引导和生命周期行为。
package host

import (
	"context"
	"errors"
	"testing"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/lifecycle"
)

// TestProvider_Name verifies that the host provider returns the correct name.
//
// TestProvider_Name 验证 host provider 返回正确的名称。
func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	if p.Name() != "host" {
		t.Errorf("expected name 'host', got %s", p.Name())
	}
}

// TestProvider_IsDefer verifies that the host provider is not deferrable.
//
// TestProvider_IsDefer 验证 host provider 不可延迟加载。
func TestProvider_IsDefer(t *testing.T) {
	p := NewProvider()
	if p.IsDefer() {
		t.Error("expected IsDefer to be false")
	}
}

// TestProvider_Provides verifies that the host provider provides the host key.
//
// TestProvider_Provides 验证 host provider 提供 host key。
func TestProvider_Provides(t *testing.T) {
	p := NewProvider()
	provides := p.Provides()
	if len(provides) != 1 || provides[0] != runtimecontract.HostKey {
		t.Errorf("expected provides [%s], got %v", runtimecontract.HostKey, provides)
	}
}

// TestDefaultHost_RegisterService verifies that services can be registered and retrieved from the default host.
//
// TestDefaultHost_RegisterService 验证服务可以正确注册到默认 host 并从中获取。
func TestDefaultHost_RegisterService(t *testing.T) {
	h := NewDefaultHost(nil)
	svc := &mockHostable{name: "test"}

	if err := h.RegisterService("test", svc); err != nil {
		t.Errorf("RegisterService failed: %v", err)
	}

	services := h.Services()
	if len(services) != 1 || services[0] != "test" {
		t.Errorf("expected services [test], got %v", services)
	}
}

// TestDefaultHost_RegisterServiceWithPriority verifies that services registered with priority are ordered correctly.
//
// TestDefaultHost_RegisterServiceWithPriority 验证带优先级的服务注册能正确排序。
func TestDefaultHost_RegisterServiceWithPriority(t *testing.T) {
	h := NewDefaultHost(nil)
	svc1 := &mockHostable{name: "svc1"}
	svc2 := &mockHostable{name: "svc2"}

	h.RegisterServiceWithPriority("svc1", svc1, nil, 100)
	h.RegisterServiceWithPriority("svc2", svc2, nil, 200)

	services := h.Services()
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
}

// TestDefaultHost_StartStop verifies that the host can start and stop services correctly.
//
// TestDefaultHost_StartStop 验证 host 能正确启动和停止服务。
func TestDefaultHost_StartStop(t *testing.T) {
	h := NewDefaultHost(nil)
	svc := &mockHostable{name: "test"}
	h.RegisterService("test", svc)

	ctx := context.Background()
	if err := h.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	if !svc.started {
		t.Error("service should be started")
	}

	if h.State() != lifecycle.StateRunning {
		t.Errorf("expected state Running, got %v", h.State())
	}

	if err := h.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	if !svc.stopped {
		t.Error("service should be stopped")
	}
}

// TestDefaultHost_Start_Idempotent verifies that starting a host multiple times only starts services once.
//
// TestDefaultHost_Start_Idempotent 验证多次启动 host 时服务只会被启动一次。
func TestDefaultHost_Start_Idempotent(t *testing.T) {
	h := NewDefaultHost(nil)
	svc := &mockHostable{name: "test"}
	h.RegisterService("test", svc)

	ctx := context.Background()
	_ = h.Start(ctx)
	_ = h.Start(ctx) // 再次调用

	// 应该只启动一次
	if svc.startCalls != 1 {
		t.Errorf("expected 1 start call, got %d", svc.startCalls)
	}
}

// TestDefaultHost_Stop_BeforeStart verifies that stopping a host before starting it does not return an error.
//
// TestDefaultHost_Stop_BeforeStart 验证在启动前停止 host 不会返回错误。
func TestDefaultHost_Stop_BeforeStart(t *testing.T) {
	h := NewDefaultHost(nil)
	svc := &mockHostable{name: "test"}
	h.RegisterService("test", svc)

	ctx := context.Background()
	// 未启动时调用 Stop 不应该报错
	if err := h.Stop(ctx); err != nil {
		t.Errorf("Stop before start should not error: %v", err)
	}
}

// mockHostable 是用于测试的 mock Hostable 实现。
type mockHostable struct {
	name       string
	startErr   error
	stopErr    error
	started    bool
	stopped    bool
	startCalls int
	stopCalls  int
}

func (m *mockHostable) Name() string { return m.name }

func (m *mockHostable) Start(ctx context.Context) error {
	m.startCalls++
	m.started = true
	return m.startErr
}

func (m *mockHostable) Stop(ctx context.Context) error {
	m.stopCalls++
	m.stopped = true
	return m.stopErr
}

// TestDefaultHost_StartError verifies that the host returns an error when a service fails to start.
//
// TestDefaultHost_StartError 验证当服务启动失败时 host 返回错误。
func TestDefaultHost_StartError(t *testing.T) {
	h := NewDefaultHost(nil)
	svc := &mockHostable{name: "test", startErr: errors.New("start error")}
	h.RegisterService("test", svc)

	ctx := context.Background()
	err := h.Start(ctx)
	if err == nil {
		t.Error("expected error from Start")
	}
}

// TestHTTPService_Name verifies that the HTTP service returns its configured name.
//
// TestHTTPService_Name 验证 HTTP 服务返回其配置的名称。
func TestHTTPService_Name(t *testing.T) {
	svc := &HTTPService{name: "http"}
	if svc.Name() != "http" {
		t.Errorf("expected name 'http', got %s", svc.Name())
	}
}

// TestCronService_Name verifies that the Cron service returns its configured name.
//
// TestCronService_Name 验证 Cron 服务返回其配置的名称。
func TestCronService_Name(t *testing.T) {
	svc := &CronService{name: "cron"}
	if svc.Name() != "cron" {
		t.Errorf("expected name 'cron', got %s", svc.Name())
	}
}
