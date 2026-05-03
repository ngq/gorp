package zookeeper

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.zookeeper", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.RPCRegistryKey}, p.Provides())
}

func TestRegistryRegisterCreatesEphemeralNode(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)

	instancePath := path.Join("/services", "user-service", "10.0.0.1:8080")
	require.Contains(t, backend.nodes, instancePath)
}

func TestRegistryDiscoverReadsInstances(t *testing.T) {
	backend := newFakeZKBackend()
	record := serviceRecord{
		ID:       "user-service-10.0.0.1:8080",
		Name:     "user-service",
		Address:  "10.0.0.1:8080",
		Metadata: map[string]string{"version": "v1"},
		Healthy:  true,
	}
	payload, err := json.Marshal(record)
	require.NoError(t, err)
	backend.nodes[path.Join("/services", "user-service", "10.0.0.1:8080")] = payload

	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "user-service")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
}

func TestRegistryDeregisterDeletesNode(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.NoError(t, err)
	require.NotContains(t, backend.nodes, path.Join("/services", "user-service", "10.0.0.1:8080"))
}

func TestRegistryDiscoverReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithBackend(testZKConfig(), newFakeZKBackend())
	require.NoError(t, err)

	_, err = registry.Discover(context.Background(), "missing-service")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryReturnsBackendError(t *testing.T) {
	expected := errors.New("zk unavailable")
	backend := newFakeZKBackend()
	backend.childrenErr = expected
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, expected)
}

func TestRegistryRegisterRejectsDuplicateInstance(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrAlreadyRegistered)
}

func TestRegistryDeregisterMissingInstanceReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithBackend(testZKConfig(), newFakeZKBackend())
	require.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryCloseRejectsOperations(t *testing.T) {
	registry, err := NewRegistryWithBackend(testZKConfig(), newFakeZKBackend())
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegistryClosed)

	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryCloseIsIdempotent(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)
	require.NoError(t, registry.Close())
	require.NoError(t, registry.Close())
	require.True(t, backend.closed)
}

func TestRegistryWatchReceivesInitialAndRemovalUpdates(t *testing.T) {
	backend := newFakeZKBackend()
	record := serviceRecord{
		ID:       "user-service-10.0.0.1:8080",
		Name:     "user-service",
		Address:  "10.0.0.1:8080",
		Metadata: map[string]string{"version": "v1"},
		Healthy:  true,
	}
	payload, err := json.Marshal(record)
	require.NoError(t, err)
	backend.nodes[path.Join("/services", "user-service", "10.0.0.1:8080")] = payload

	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.1:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	require.NoError(t, backend.Delete(path.Join("/services", "user-service", "10.0.0.1:8080")))

	select {
	case instances := <-ch:
		require.Len(t, instances, 0)
	case <-time.After(2 * time.Second):
		t.Fatal("expected empty snapshot after node removal")
	}
}

func TestRegistryWatchAfterCloseFails(t *testing.T) {
	registry, err := NewRegistryWithBackend(testZKConfig(), newFakeZKBackend())
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	_, err = registry.Watch(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchRetriesAfterBackendError(t *testing.T) {
	backend := newFakeZKBackend()
	backend.watchErrs = []error{zk.ErrConnectionClosed}
	record := serviceRecord{
		ID:       "user-service-10.0.0.1:8080",
		Name:     "user-service",
		Address:  "10.0.0.1:8080",
		Metadata: map[string]string{"version": "v1"},
		Healthy:  true,
	}
	payload, err := json.Marshal(record)
	require.NoError(t, err)
	backend.nodes[path.Join("/services", "user-service", "10.0.0.1:8080")] = payload

	cfg := testZKConfig()
	cfg.WatchRetryInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithBackend(cfg, backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	require.Eventually(t, func() bool {
		return backend.watchCalls >= 2
	}, time.Second, 10*time.Millisecond)

	require.NoError(t, backend.Delete(path.Join("/services", "user-service", "10.0.0.1:8080")))

	select {
	case instances := <-ch:
		require.Len(t, instances, 0)
	case <-time.After(2 * time.Second):
		t.Fatal("expected empty snapshot after retry")
	}
	require.GreaterOrEqual(t, backend.watchCalls, 2)
}

func TestRegistryWatchRetriesAfterSessionExpired(t *testing.T) {
	backend := newFakeZKBackend()
	backend.watchErrs = []error{zk.ErrSessionExpired}
	record := serviceRecord{
		ID:       "user-service-10.0.0.1:8080",
		Name:     "user-service",
		Address:  "10.0.0.1:8080",
		Metadata: map[string]string{"version": "v1"},
		Healthy:  true,
	}
	payload, err := json.Marshal(record)
	require.NoError(t, err)
	backend.nodes[path.Join("/services", "user-service", "10.0.0.1:8080")] = payload

	cfg := testZKConfig()
	cfg.WatchRetryInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithBackend(cfg, backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	require.Eventually(t, func() bool {
		return backend.watchCalls >= 2
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryWatchStopsOnNonRetryableError(t *testing.T) {
	backend := newFakeZKBackend()
	backend.watchErrs = []error{errors.New("zk auth failed")}
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case instances, ok := <-ch:
		if ok {
			require.Empty(t, instances)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch channel to close")
	}

	require.Eventually(t, func() bool {
		return backend.watchCalls == 1
	}, time.Second, 10*time.Millisecond)

	select {
	case _, ok := <-ch:
		require.False(t, ok)
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch channel to be closed")
	}
}

func TestRegistryWatchSkipsDuplicateEmptySnapshot(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case instances := <-ch:
		require.Empty(t, instances)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial empty snapshot")
	}

	backend.TriggerWatch(path.Join("/services", "user-service"))

	select {
	case instances := <-ch:
		t.Fatalf("expected duplicate empty snapshot to be suppressed, got %#v", instances)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestRegistryWatchChannelClosesAfterRegistryClose(t *testing.T) {
	backend := newFakeZKBackend()
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	require.NoError(t, registry.Close())

	select {
	case _, ok := <-ch:
		require.False(t, ok)
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch channel to close after registry close")
	}
}

func TestRegistryUnderlyingAndAs(t *testing.T) {
	backend := &fakeNativeZKBackend{
		fakeZKBackend: newFakeZKBackend(),
		native:        "native-zk",
	}
	registry, err := NewRegistryWithBackend(testZKConfig(), backend)
	require.NoError(t, err)

	require.Equal(t, "native-zk", registry.Underlying())

	var projected string
	require.True(t, registry.As(&projected))
	require.Equal(t, "native-zk", projected)
}

type fakeZKBackend struct {
	mu          sync.RWMutex
	nodes       map[string][]byte
	childrenErr error
	getErr      error
	createErr   error
	deleteErr   error
	closed      bool
	watchers    map[string][]chan struct{}
	watchErrs   []error
	watchCalls  int
}

type fakeNativeZKBackend struct {
	*fakeZKBackend
	native any
}

func (b *fakeNativeZKBackend) Underlying() any {
	return b.native
}

func newFakeZKBackend() *fakeZKBackend {
	return &fakeZKBackend{
		nodes:    make(map[string][]byte),
		watchers: make(map[string][]chan struct{}),
	}
}

func (b *fakeZKBackend) EnsurePath(target string) error {
	return nil
}

func (b *fakeZKBackend) CreateEphemeral(target string, data []byte) error {
	if b.createErr != nil {
		return b.createErr
	}
	b.mu.Lock()
	b.nodes[target] = data
	b.notifyWatchersLocked(parentPath(target))
	b.mu.Unlock()
	return nil
}

func (b *fakeZKBackend) Delete(target string) error {
	if b.deleteErr != nil {
		return b.deleteErr
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.nodes[target]; !ok {
		return zk.ErrNoNode
	}
	delete(b.nodes, target)
	b.notifyWatchersLocked(parentPath(target))
	return nil
}

func (b *fakeZKBackend) Children(target string) ([]string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.childrenErr != nil {
		return nil, b.childrenErr
	}
	children := make([]string, 0)
	prefix := stringsTrimSuffix(target, "/") + "/"
	for node := range b.nodes {
		if stringsHasPrefix(node, prefix) {
			children = append(children, node[len(prefix):])
		}
	}
	if len(children) == 0 {
		return nil, zk.ErrNoNode
	}
	return children, nil
}

func (b *fakeZKBackend) Get(target string) ([]byte, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.getErr != nil {
		return nil, b.getErr
	}
	data, ok := b.nodes[target]
	if !ok {
		return nil, zk.ErrNoNode
	}
	return data, nil
}

func (b *fakeZKBackend) WatchChildren(ctx context.Context, target string, onUpdate func()) error {
	b.mu.Lock()
	b.watchCalls++
	if len(b.watchErrs) > 0 {
		err := b.watchErrs[0]
		b.watchErrs = b.watchErrs[1:]
		b.mu.Unlock()
		return err
	}
	ch := make(chan struct{}, 4)
	b.watchers[target] = append(b.watchers[target], ch)
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		watchers := b.watchers[target]
		for i, watcher := range watchers {
			if watcher == ch {
				b.watchers[target] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(b.watchers[target]) == 0 {
			delete(b.watchers, target)
		}
		b.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ch:
			onUpdate()
		}
	}
}

func (b *fakeZKBackend) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	b.nodes = make(map[string][]byte)
	for target := range b.watchers {
		b.notifyWatchersLocked(target)
	}
	return nil
}

func (b *fakeZKBackend) TriggerWatch(target string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.notifyWatchersLocked(target)
}

func testZKConfig() *ZookeeperConfig {
	return &ZookeeperConfig{
		Servers:            []string{"127.0.0.1:2181"},
		BasePath:           "/services",
		SessionTimeout:     time.Second,
		WatchRetryInterval: 200 * time.Millisecond,
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsTrimSuffix(s, suffix string) string {
	for len(s) > 0 && len(suffix) > 0 && s[len(s)-1] == suffix[len(suffix)-1] {
		return s[:len(s)-1]
	}
	return s
}

func parentPath(target string) string {
	if idx := len(target) - 1; idx >= 0 {
		return path.Dir(target)
	}
	return target
}

func (b *fakeZKBackend) notifyWatchersLocked(target string) {
	watchers := append([]chan struct{}(nil), b.watchers[target]...)
	for _, watcher := range watchers {
		select {
		case watcher <- struct{}{}:
		default:
		}
	}
}
