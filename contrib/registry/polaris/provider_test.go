package polaris

import (
	"context"
	"errors"
	"testing"
	"time"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.polaris", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{transportcontract.RPCRegistryKey}, p.Provides())
}

func TestRegistryRegisterUsesClient(t *testing.T) {
	client := &fakePolarisRegistryClient{}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)
	require.Equal(t, 1, client.registerCalls)
}

func TestRegistryDeregisterUsesClient(t *testing.T) {
	client := &fakePolarisRegistryClient{}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.NoError(t, err)
	require.Equal(t, 1, client.deregisterCalls)
}

func TestRegistryRegisterRejectsDuplicateInstance(t *testing.T) {
	client := &fakePolarisRegistryClient{}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrAlreadyRegistered)
}

func TestRegistryDiscoverUsesClient(t *testing.T) {
	client := &fakePolarisRegistryClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "user-service")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
}

func TestRegistryDiscoverReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), &fakePolarisRegistryClient{
		discoverErr: ErrServiceNotFound,
	})
	require.NoError(t, err)
	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryUnderlyingReturnsClient(t *testing.T) {
	client := &fakePolarisRegistryClient{}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	require.Same(t, client, registry.Underlying())
}

func TestRegistryAsProjectsClient(t *testing.T) {
	client := &fakePolarisRegistryClient{}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	var projected *fakePolarisRegistryClient
	require.True(t, registry.As(&projected))
	require.Same(t, client, projected)
}

func TestRegistryRegisterReturnsSourceError(t *testing.T) {
	expected := errors.New("polaris unavailable")
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), &fakePolarisRegistryClient{
		registerErr: expected,
	})
	require.NoError(t, err)
	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, expected)
}

func TestRegistryCloseRejectsOperations(t *testing.T) {
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), &fakePolarisRegistryClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())
	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegistryClosed)
	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchDeliversInitialAndUpdatedSnapshot(t *testing.T) {
	client := &fakePolarisRegistryClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchUpdates: make(chan []transportcontract.ServiceInstance, 1),
	}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
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

	client.push([]transportcontract.ServiceInstance{
		{ID: "user-service-10.0.0.2:8080", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.2:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated watch snapshot")
	}
}

func TestRegistryWatchAfterCloseFails(t *testing.T) {
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), &fakePolarisRegistryClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	_, err = registry.Watch(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchRetriesAfterClientError(t *testing.T) {
	client := &fakePolarisRegistryClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "user-service-10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchErrs:    []error{errors.New("polaris watch unavailable")},
		watchUpdates: make(chan []transportcontract.ServiceInstance, 1),
	}
	cfg := testPolarisRegistryConfig()
	cfg.WatchRetryInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial snapshot")
	}

	client.push([]transportcontract.ServiceInstance{
		{ID: "user-service-10.0.0.9:8080", Name: "user-service", Address: "10.0.0.9:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		require.Equal(t, "10.0.0.9:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected update after watch retry")
	}
	require.GreaterOrEqual(t, client.watchCalls, 2)
}

func TestRegistryWatchSkipsDuplicateSnapshotWithDifferentOrder(t *testing.T) {
	client := &fakePolarisRegistryClient{
		discoverResult: []transportcontract.ServiceInstance{
			{ID: "b", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
			{ID: "a", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchUpdates: make(chan []transportcontract.ServiceInstance, 2),
	}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case instances := <-ch:
		require.Len(t, instances, 2)
		require.Equal(t, "a", instances[0].ID)
		require.Equal(t, "b", instances[1].ID)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	client.push([]transportcontract.ServiceInstance{
		{ID: "a", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		{ID: "b", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		t.Fatalf("unexpected duplicate snapshot: %+v", instances)
	case <-time.After(150 * time.Millisecond):
	}
}

type fakePolarisRegistryClient struct {
	registerErr     error
	deregisterErr   error
	discoverErr     error
	discoverResult  []transportcontract.ServiceInstance
	watchUpdates    chan []transportcontract.ServiceInstance
	watchErrs       []error
	registerCalls   int
	deregisterCalls int
	discoverCalls   int
	watchCalls      int
}

type fakePolarisRegistryNativeClient struct {
	fakePolarisRegistryClient
	native any
}

func (f *fakePolarisRegistryNativeClient) Underlying() any {
	return f.native
}

func (f *fakePolarisRegistryClient) Register(ctx context.Context, cfg *PolarisConfig, name, addr string, meta map[string]string) error {
	f.registerCalls++
	return f.registerErr
}

func (f *fakePolarisRegistryClient) Deregister(ctx context.Context, cfg *PolarisConfig, name, addr string) error {
	f.deregisterCalls++
	return f.deregisterErr
}

func (f *fakePolarisRegistryClient) Discover(ctx context.Context, cfg *PolarisConfig, name string) ([]transportcontract.ServiceInstance, error) {
	f.discoverCalls++
	if f.discoverErr != nil {
		return nil, f.discoverErr
	}
	return append([]transportcontract.ServiceInstance(nil), f.discoverResult...), nil
}

func (f *fakePolarisRegistryClient) Watch(ctx context.Context, cfg *PolarisConfig, name string, onUpdate func([]transportcontract.ServiceInstance)) error {
	f.watchCalls++
	if len(f.watchErrs) > 0 {
		err := f.watchErrs[0]
		f.watchErrs = f.watchErrs[1:]
		return err
	}
	if f.watchUpdates == nil {
		<-ctx.Done()
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case instances := <-f.watchUpdates:
			onUpdate(instances)
		}
	}
}

func (f *fakePolarisRegistryClient) push(instances []transportcontract.ServiceInstance) {
	if f.watchUpdates == nil {
		f.watchUpdates = make(chan []transportcontract.ServiceInstance, 1)
	}
	f.watchUpdates <- instances
}

func testPolarisRegistryConfig() *PolarisConfig {
	return &PolarisConfig{Address: "http://polaris.local", Namespace: "default", WatchRetryInterval: 200 * time.Millisecond}
}

func TestRegistryUnderlyingPrefersNativeClient(t *testing.T) {
	native := &struct{ Name string }{Name: "registry-native"}
	registry, err := NewRegistryWithClient(testPolarisRegistryConfig(), &fakePolarisRegistryNativeClient{native: native})
	require.NoError(t, err)

	require.Same(t, native, registry.Underlying())
}
