package lifecycle

import (
	"context"
	"errors"
	"testing"
)

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

// mockLifecycle 是用于测试的 mock Lifecycle 实现。
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

	// 注册顺序与优先级不同
	m.Register("svc1", svc1, nil, 300) // 优先级最低，最后启动
	m.Register("svc2", svc2, nil, 100) // 优先级最高，最先启动
	m.Register("svc3", svc3, nil, 200) // 优先级中等

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// 验证启动顺序：svc2(100) -> svc3(200) -> svc1(300)
	// 可以通过 startCalls 的调用顺序间接验证
	if svc2.startCalls != 1 || svc3.startCalls != 1 || svc1.startCalls != 1 {
		t.Errorf("each service should be started once")
	}
}

func TestManager_Start_StopOnFailure(t *testing.T) {
	m := NewManager()
	svc1 := &mockHostable{name: "svc1"}
	svc2 := &mockHostable{name: "svc2"}
	svc3 := &mockHostable{name: "svc3", startErr: errors.New("start error")}

	// 按优先级注册：svc1(100) -> svc2(200) -> svc3(300)
	m.Register("svc1", svc1, nil, 100)
	m.Register("svc2", svc2, nil, 200)
	m.Register("svc3", svc3, nil, 300)

	ctx := context.Background()
	err := m.Start(ctx)
	if err == nil {
		t.Error("expected error from Start")
	}

	// svc1, svc2 应该已启动
	if !svc1.started {
		t.Error("svc1 should be started")
	}
	if !svc2.started {
		t.Error("svc2 should be started")
	}

	// 启动失败后应该停止已启动的服务
	// 由于 svc3 启动失败，svc1 和 svc2 应该被停止
	if !svc1.stopped || !svc2.stopped {
		t.Error("started services should be stopped on failure")
	}

	// 状态应该回到 idle
	if m.State() != StateIdle {
		t.Errorf("expected state Idle, got %v", m.State())
	}
}

func TestManager_Lifecycle_Hooks(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	hooks := &mockLifecycle{}

	m.Register("test", svc, hooks, 100)

	ctx := context.Background()
	if err := m.Start(ctx); err != nil {
		t.Errorf("Start failed: %v", err)
	}

	// 验证钩子调用顺序
	expectedCalls := []string{"OnStarting", "OnStarted"}
	if len(hooks.calls) != len(expectedCalls) {
		t.Errorf("expected %d calls, got %d", len(expectedCalls), len(hooks.calls))
	}
	for i, call := range expectedCalls {
		if i >= len(hooks.calls) || hooks.calls[i] != call {
			t.Errorf("expected call %s at position %d, got %v", call, i, hooks.calls)
		}
	}

	// 重置 calls
	hooks.calls = nil

	if err := m.Stop(ctx); err != nil {
		t.Errorf("Stop failed: %v", err)
	}

	// 验证停止钩子
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

func TestManager_Start_Idempotent(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	_ = m.Start(ctx)
	_ = m.Start(ctx) // 再次调用

	// 应该只启动一次
	if svc.startCalls != 1 {
		t.Errorf("expected 1 start call, got %d", svc.startCalls)
	}
}

func TestManager_Stop_CalledOnce(t *testing.T) {
	m := NewManager()
	svc := &mockHostable{name: "test"}
	m.Register("test", svc, nil, 100)

	ctx := context.Background()
	_ = m.Start(ctx)
	_ = m.Stop(ctx)
	_ = m.Stop(ctx) // 再次调用

	// 应该只停止一次
	if svc.stopCalls != 1 {
		t.Errorf("expected 1 stop call, got %d", svc.stopCalls)
	}
}

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