package servicecomb

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.servicecomb", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{transportcontract.RPCRegistryKey}, p.Provides())
}

func TestRegistryRegisterUsesClient(t *testing.T) {
	client := &fakeServiceCombClient{}
	registry, err := NewRegistryWithClient(testServiceCombConfig(), client)
	require.NoError(t, err)
	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"region": "cn"})
	require.NoError(t, err)
	require.Equal(t, int32(1), client.registerCalls.Load())
}

func TestRegistryDeregisterUsesClient(t *testing.T) {
	client := &fakeServiceCombClient{}
	registry, err := NewRegistryWithClient(testServiceCombConfig(), client)
	require.NoError(t, err)
	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.NoError(t, err)
	require.Equal(t, int32(1), client.deregisterCalls.Load())
}

func TestRegistryRegisterRejectsDuplicateInstance(t *testing.T) {
	client := &fakeServiceCombClient{}
	cfg := testServiceCombConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrAlreadyRegistered)
}

func TestRegistryDiscoverUsesClient(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(testServiceCombConfig(), client)
	require.NoError(t, err)
	instances, err := registry.Discover(context.Background(), "user-service")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
}

func TestRegistryDiscoverReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithClient(testServiceCombConfig(), &fakeServiceCombClient{discoverErr: ErrServiceNotFound})
	require.NoError(t, err)
	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryUnderlyingAndAs(t *testing.T) {
	client := &fakeNativeServiceCombClient{
		fakeServiceCombClient: &fakeServiceCombClient{},
		native:                "native-servicecomb",
	}
	registry, err := NewRegistryWithClient(testServiceCombConfig(), client)
	require.NoError(t, err)

	require.Equal(t, "native-servicecomb", registry.Underlying())

	var projected string
	require.True(t, registry.As(&projected))
	require.Equal(t, "native-servicecomb", projected)
}

func TestRegistryRegisterReturnsSourceError(t *testing.T) {
	expected := errors.New("servicecenter unavailable")
	registry, err := NewRegistryWithClient(testServiceCombConfig(), &fakeServiceCombClient{registerErr: expected})
	require.NoError(t, err)
	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, expected)
}

func TestRegistryCloseRejectsOperations(t *testing.T) {
	registry, err := NewRegistryWithClient(testServiceCombConfig(), &fakeServiceCombClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())
	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegistryClosed)
	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryHeartbeatRunsAfterRegister(t *testing.T) {
	client := &fakeServiceCombClient{}
	cfg := testServiceCombConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.heartbeatCalls.Load() > 0
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryDeregisterStopsHeartbeat(t *testing.T) {
	client := &fakeServiceCombClient{}
	cfg := testServiceCombConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return client.heartbeatCalls.Load() > 0
	}, time.Second, 10*time.Millisecond)

	before := client.heartbeatCalls.Load()
	require.NoError(t, registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080"))
	time.Sleep(40 * time.Millisecond)
	require.Equal(t, before, client.heartbeatCalls.Load())
}

func TestRegistryHeartbeatNotFoundTriggersReRegister(t *testing.T) {
	client := &fakeServiceCombClient{
		heartbeatErrs: []error{ErrServiceNotFound, nil},
	}
	cfg := testServiceCombConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatRetryBackoff = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"region": "cn"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.registerCalls.Load() >= 2
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryWatchDeliversInitialAndUpdatedSnapshot(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverResults: [][]transportcontract.ServiceInstance{
			{
				{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
			},
			{
				{ID: "user-service-10.0.0.2:8080", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
			},
		},
	}
	cfg := testServiceCombConfig()
	cfg.HeartbeatRetryBackoff = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
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

	require.Eventually(t, func() bool {
		select {
		case instances := <-ch:
			return len(instances) == 1 && instances[0].Address == "10.0.0.2:8080"
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryWatchSkipsDuplicateSnapshot(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverResults: [][]transportcontract.ServiceInstance{
			{
				{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
			},
			{
				{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
			},
		},
	}
	cfg := testServiceCombConfig()
	cfg.WatchInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	select {
	case instances := <-ch:
		t.Fatalf("expected duplicate snapshot to be suppressed, got %#v", instances)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestRegistryWatchAfterCloseFails(t *testing.T) {
	registry, err := NewRegistryWithClient(testServiceCombConfig(), &fakeServiceCombClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	_, err = registry.Watch(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchRetriesAfterSourceError(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverErrs: []error{
			nil,
			errors.New("temporary unavailable"),
			nil,
		},
		discoverResults: [][]transportcontract.ServiceInstance{
			{
				{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
			},
			{
				{ID: "user-service-10.0.0.2:8080", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
			},
		},
	}
	cfg := testServiceCombConfig()
	cfg.WatchInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
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

	require.Eventually(t, func() bool {
		select {
		case instances := <-ch:
			return len(instances) == 1 && instances[0].Address == "10.0.0.2:8080"
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryWatchEmitsEmptySnapshotAfterServiceRemoved(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverErrs: []error{
			nil,
			ErrServiceNotFound,
		},
		discoverResults: [][]transportcontract.ServiceInstance{
			{
				{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
			},
		},
	}
	cfg := testServiceCombConfig()
	cfg.WatchInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	require.Eventually(t, func() bool {
		select {
		case instances := <-ch:
			return len(instances) == 0
		default:
			return false
		}
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryWatchChannelClosesAfterRegistryClose(t *testing.T) {
	client := &fakeServiceCombClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	cfg := testServiceCombConfig()
	cfg.WatchInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
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

type fakeServiceCombClient struct {
	registerErr     error
	deregisterErr   error
	heartbeatErr    error
	heartbeatErrs   []error
	discoverErr     error
	discoverResult  []transportcontract.ServiceInstance
	discoverErrs    []error
	discoverResults [][]transportcontract.ServiceInstance
	registerCalls   atomic.Int32
	deregisterCalls atomic.Int32
	heartbeatCalls  atomic.Int32
	discoverCalls   atomic.Int32
}

type fakeNativeServiceCombClient struct {
	*fakeServiceCombClient
	native any
}

func (f *fakeNativeServiceCombClient) Underlying() any {
	return f.native
}

func (f *fakeServiceCombClient) Register(ctx context.Context, cfg *ServiceCombConfig, name, addr string, meta map[string]string) error {
	f.registerCalls.Add(1)
	return f.registerErr
}

func (f *fakeServiceCombClient) Deregister(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	f.deregisterCalls.Add(1)
	return f.deregisterErr
}

func (f *fakeServiceCombClient) Heartbeat(ctx context.Context, cfg *ServiceCombConfig, name, addr string) error {
	f.heartbeatCalls.Add(1)
	if len(f.heartbeatErrs) > 0 {
		err := f.heartbeatErrs[0]
		f.heartbeatErrs = f.heartbeatErrs[1:]
		return err
	}
	return f.heartbeatErr
}

func (f *fakeServiceCombClient) Discover(ctx context.Context, cfg *ServiceCombConfig, name string) ([]transportcontract.ServiceInstance, error) {
	f.discoverCalls.Add(1)
	if len(f.discoverErrs) > 0 {
		err := f.discoverErrs[0]
		f.discoverErrs = f.discoverErrs[1:]
		if err != nil {
			return nil, err
		}
	}
	if f.discoverErr != nil {
		return nil, f.discoverErr
	}
	if len(f.discoverResults) > 0 {
		result := f.discoverResults[0]
		if len(f.discoverResults) > 1 {
			f.discoverResults = f.discoverResults[1:]
		}
		return append([]transportcontract.ServiceInstance(nil), result...), nil
	}
	return append([]transportcontract.ServiceInstance(nil), f.discoverResult...), nil
}

func testServiceCombConfig() *ServiceCombConfig {
	return &ServiceCombConfig{
		ServerURI:             "http://servicecenter.local",
		AppID:                 "demo",
		Version:               "1.0.0",
		Environment:           "production",
		HeartbeatInterval:     0,
		HeartbeatRetryBackoff: time.Second,
	}
}
