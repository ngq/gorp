package lifecycle_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract/transport"
	"github.com/ngq/gorp/framework/lifecycle"
	"github.com/stretchr/testify/require"
)

type mockRegistry struct {
	mu          sync.Mutex
	registered  map[string]string
	deregistered map[string]string
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		registered:   make(map[string]string),
		deregistered: make(map[string]string),
	}
}

func (m *mockRegistry) Register(ctx context.Context, name, addr string, meta map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registered[name] = addr
	return nil
}

func (m *mockRegistry) Deregister(ctx context.Context, name, addr string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deregistered[name] = addr
	delete(m.registered, name)
	return nil
}

func (m *mockRegistry) Discover(ctx context.Context, name string) ([]transport.ServiceInstance, error) {
	return nil, nil
}

func (m *mockRegistry) Close() error { return nil }

func TestRegistryHook_PostStartAndPreStopOrder(t *testing.T) {
	reg := newMockRegistry()
	hook := lifecycle.NewRegistryHook(reg, lifecycle.RegistryHookOptions{
		ServiceName: "user-service",
		Address:     "127.0.0.1:8080",
		Metadata:    map[string]string{"version": "v1"},
		DrainDelay:  10 * time.Millisecond,
	})

	ctx := context.Background()

	// 1. Post-Start registration
	err := hook.OnStarted(ctx)
	require.NoError(t, err)

	reg.mu.Lock()
	require.Equal(t, "127.0.0.1:8080", reg.registered["user-service"])
	reg.mu.Unlock()

	// 2. Pre-Stop deregistration
	start := time.Now()
	err = hook.OnStopping(ctx)
	require.NoError(t, err)
	elapsed := time.Since(start)

	require.GreaterOrEqual(t, elapsed, 10*time.Millisecond)

	reg.mu.Lock()
	require.Equal(t, "127.0.0.1:8080", reg.deregistered["user-service"])
	require.Empty(t, reg.registered["user-service"])
	reg.mu.Unlock()
}
