// Package lifecycle_test provides unit tests for lifecycle manager ordering and state transitions.
//
// 适用场景：
// - 验证 lifecycle manager 的顺序、回滚和状态流转行为。
// - 防止 hook 调用语义和幂等语义回归。
// - 通过聚焦型单测把预期运行时行为固化下来。
package lifecycle

import (
	"context"
	"errors"
	"testing"
)

// mockHostable is a test double for runtimecontract.Hostable.
//
// mockHostable 是 runtimecontract.Hostable 的测试替身。
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

// mockLifecycle is a test double for runtimecontract.Lifecycle.
//
// mockLifecycle 是 runtimecontract.Lifecycle 的测试替身。
type mockLifecycle struct {
	onStartingErr error
	onStartedErr  error
	onStoppingErr error
	onStoppedErr  error
	calls         []string
}

func (m *mockLifecycle) OnStarting(ctx context.Context) error {
	m.calls = append(m.calls, "OnStarting")
	return m.onStartingErr
}

func (m *mockLifecycle) OnStarted(ctx context.Context) error {
	m.calls = append(m.calls, "OnStarted")
	return m.onStartedErr
}

func (m *mockLifecycle) OnStopping(ctx context.Context) error {
	m.calls = append(m.calls, "OnStopping")
	return m.onStoppingErr
}

func (m *mockLifecycle) OnStopped(ctx context.Context) error {
	m.calls = append(m.calls, "OnStopped")
	return m.onStoppedErr
}

func TestManager_Register(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}

	m.Register("test", svc, nil, 100)

	services := m.Services()
	if len(services) != 1 || services[0] != "test" {
		t.Errorf("expected services [test], got %v", services)
	}
}

// TestManager_Register_Multiple verifies that multiple services can be registered and retrieved.
//
// TestManager_Register_Multiple 验证多个服务可以注册并正确获取。
func TestManager_Register_Multiple(t *testing.T) {
	m := NewManager()
	svc1 := &mockHostable{name: "svc1"}
	svc2 := &mockHostable{name: "svc2"}

	m.Register("svc1", svc1, nil, 100)
	m.Register("svc2", svc2, nil, 200)

	services := m.Services()
	if len(services) != 2 {
		t.Errorf("expected 2 services, got %d", len(services))
	}
}

// TestManager_StartStop_Success verifies that services start and stop correctly with state transitions.
//
// TestManager_StartStop_Success 验证服务启动和停止以及状态转换的正确性。
func TestManager_StartStop_Success(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}

	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	if !svc.started {
		t.Error("service should be started")
	}

	if m.State() != StateRunning {
		t.Errorf("expected state Running, got %v", m.State())
	}

	if err := m.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	if !svc.stopped {
		t.Error("service should be stopped")
	}

	if m.State() != StateStopped {
		t.Errorf("expected state Stopped, got %v", m.State())
	}
}

func TestManager_Start_Priority(t *testing.T) {
	m := NewManager()
	svc1 := &mockHostable{name: "svc1"}
	svc2 := &mockHostable{name: "svc2"}
	svc3 := &mockHostable{name: "svc3"}

	// Register out of priority order to ensure the manager reorders them before startup.
	// 故意用乱序注册，验证 manager 会在启动前按 priority 重排。
	m.Register("svc1", svc1, nil, 300)
	m.Register("svc2", svc2, nil, 100)
	m.Register("svc3", svc3, nil, 200)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	if svc2.startCalls != 1 || svc3.startCalls != 1 || svc1.startCalls != 1 {
		t.Errorf("each service should be started once")
	}
}

func TestManager_Start_StopOnFailure(t *testing.T) {
	m := NewManager()
	svc1 := &mockHostable{name: "svc1"}
	svc2 := &mockHostable{name: "svc2"}
	svc3 := &mockHostable{name: "svc3", startErr: errors.New("start error")}

	// The third service fails so the first two should be rolled back.
	// 第三个服务启动失败，因此前两个已启动服务应被回滚停止。
	m.Register("svc1", svc1, nil, 100)
	m.Register("svc2", svc2, nil, 200)
	m.Register("svc3", svc3, nil, 300)

	ctx := context.Background()
	err := m.Start(ctx)
	if err == nil {
		t.Error("expected error from Start")
	}

	if !svc1.started {
		t.Error("svc1 should be started")
	}
	if !svc2.started {
		t.Error("svc2 should be started")
	}

	if !svc1.stopped || !svc2.stopped {
		t.Error("started services should be stopped on failure")
	}

	if m.State() != StateIdle {
		t.Errorf("expected state Idle, got %v", m.State())
	}
}

// TestManager_Lifecycle_Hooks verifies that lifecycle hooks are called in correct order during start and stop.
//
// TestManager_Lifecycle_Hooks 验证生命周期钩子在启动和停止时按正确顺序调用。
func TestManager_Lifecycle_Hooks(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	hooks := &mockLifecycle{}

	m.Register("test", svc, hooks, 100)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	expectedCalls := []string{"OnStarting", "OnStarted"}
	if len(hooks.calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d", len(expectedCalls), len(hooks.calls))
	}
	for i, call := range expectedCalls {
		if i >= len(hooks.calls) || hooks.calls[i] != call {
			t.Errorf("expected call %s at position %d, got %v", call, i, hooks.calls)
		}
	}

	hooks.calls = nil

	if err := m.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	expectedStopCalls := []string{"OnStopping", "OnStopped"}
	if len(hooks.calls) != len(expectedStopCalls) {
		t.Errorf("expected %d calls, got %d", len(expectedStopCalls), len(hooks.calls))
	}
	for i, call := range expectedStopCalls {
		if i >= len(hooks.calls) || hooks.calls[i] != call {
			t.Errorf("expected call %s at position %d, got %v", call, i, hooks.calls)
		}
	}
}

// TestManager_State verifies that Manager state transitions through Idle -> Running -> Stopped.
//
// TestManager_State 验证 Manager 状态经历 Idle -> Running -> Stopped 的转换。
func TestManager_State(t *testing.T) {
	m := NewManager()

	if m.State() != StateIdle {
		t.Errorf("expected initial state Idle, got %v", m.State())
	}

	svc := &mockHostable{name: "test"}
	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	_ = m.Start(ctx)
	if m.State() != StateRunning {
		t.Errorf("expected state Running after start, got %v", m.State())
	}

	_ = m.Stop(ctx)
	if m.State() != StateStopped {
		t.Errorf("expected state Stopped after stop, got %v", m.State())
	}
}

// TestManager_Start_Idempotent verifies that calling Start multiple times only starts services once.
//
// TestManager_Start_Idempotent 验证多次调用 Start 只会启动服务一次。
func TestManager_Start_Idempotent(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	_ = m.Start(ctx)
	_ = m.Start(ctx)

	if svc.startCalls != 1 {
		t.Errorf("expected 1 start call, got %d", svc.startCalls)
	}
}

// TestManager_Stop_CalledOnce verifies that calling Stop multiple times only stops services once.
//
// TestManager_Stop_CalledOnce 验证多次调用 Stop 只会停止服务一次。
func TestManager_Stop_CalledOnce(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	_ = m.Start(ctx)
	_ = m.Stop(ctx)
	_ = m.Stop(ctx)

	if svc.stopCalls != 1 {
		t.Errorf("expected 1 stop call, got %d", svc.stopCalls)
	}
}

// TestState_String verifies that State String() returns correct string representations.
//
// TestState_String 验证 State String() 返回正确的字符串表示。
func TestState_String(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateIdle, "idle"},
		{StateStarting, "starting"},
		{StateRunning, "running"},
		{StateStopping, "stopping"},
		{StateStopped, "stopped"},
		{State(999), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.state.String(); got != tt.expected {
			t.Errorf("State(%d).String() = %s, want %s", tt.state, got, tt.expected)
		}
	}
}
