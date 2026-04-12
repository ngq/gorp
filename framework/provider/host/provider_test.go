package host

import (
	"context"
	"errors"
	"testing"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/lifecycle"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	if p.Name() != "host" {
		t.Errorf("expected name 'host', got %s", p.Name())
	}
}

func TestProvider_IsDefer(t *testing.T) {
	p := NewProvider()
	if p.IsDefer() {
		t.Error("expected IsDefer to be false")
	}
}

func TestProvider_Provides(t *testing.T) {
	p := NewProvider()
	provides := p.Provides()
	if len(provides) != 1 || provides[0] != contract.HostKey {
		t.Errorf("expected provides [%s], got %v", contract.HostKey, provides)
	}
}

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

func TestHTTPService_Name(t *testing.T) {
	svc := &HTTPService{name: "http"}
	if svc.Name() != "http" {
		t.Errorf("expected name 'http', got %s", svc.Name())
	}
}

func TestCronService_Name(t *testing.T) {
	svc := &CronService{name: "cron"}
	if svc.Name() != "cron" {
		t.Errorf("expected name 'cron', got %s", svc.Name())
	}
}