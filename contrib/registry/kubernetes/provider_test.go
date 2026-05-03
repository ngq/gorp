package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes/fake"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.kubernetes", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.RPCRegistryKey}, p.Provides())
}

func TestRegistryDiscoverUsesDiscoveryClient(t *testing.T) {
	client := &fakeDiscoveryClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "svc-10.0.0.1:8080", Name: "svc", Address: "10.0.0.1:8080", Healthy: true},
		},
	}

	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "svc")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
}

func TestRegistryWatchEmitsUpdatedInstances(t *testing.T) {
	client := &fakeDiscoveryClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "svc-10.0.0.1:8080", Name: "svc", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "svc")
	require.NoError(t, err)

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.1:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial snapshot")
	}

	client.push([]contract.ServiceInstance{
		{ID: "svc-10.0.0.2:8080", Name: "svc", Address: "10.0.0.2:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.2:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch update")
	}
}

func TestRegistryDiscoverReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, &fakeDiscoveryClient{
		discoverErr: ErrServiceNotFound,
	})
	require.NoError(t, err)

	_, err = registry.Discover(context.Background(), "svc")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryDiscoverReturnsSourceError(t *testing.T) {
	expected := errors.New("kubernetes: api unreachable")
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, &fakeDiscoveryClient{
		discoverErr: expected,
	})
	require.NoError(t, err)

	_, err = registry.Discover(context.Background(), "svc")
	require.ErrorIs(t, err, expected)
}

func TestRegistryDiscoverUsesCacheAfterFirstFetch(t *testing.T) {
	client := &fakeDiscoveryClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "svc-10.0.0.1:8080", Name: "svc", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	first, err := registry.Discover(context.Background(), "svc")
	require.NoError(t, err)
	require.Len(t, first, 1)

	client.discoverResult = []contract.ServiceInstance{
		{ID: "svc-10.0.0.9:8080", Name: "svc", Address: "10.0.0.9:8080", Healthy: true},
	}

	second, err := registry.Discover(context.Background(), "svc")
	require.NoError(t, err)
	require.Len(t, second, 1)
	require.Equal(t, "10.0.0.1:8080", second[0].Address)
	require.Equal(t, 1, client.discoverCalls)
}

func TestRegistryWatchAfterCloseFails(t *testing.T) {
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, &fakeDiscoveryClient{})
	require.NoError(t, err)

	require.NoError(t, registry.Close())

	_, err = registry.Watch(context.Background(), "svc")
	require.EqualError(t, err, "kubernetes: registry closed")
}

func TestRegistryWatchClosesChannelOnClose(t *testing.T) {
	client := &fakeDiscoveryClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "svc-10.0.0.1:8080", Name: "svc", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "svc")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial snapshot")
	}

	require.NoError(t, registry.Close())

	select {
	case _, ok := <-ch:
		require.False(t, ok)
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch channel closed after registry close")
	}
}

func TestRegistryCloseIsIdempotent(t *testing.T) {
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, &fakeDiscoveryClient{})
	require.NoError(t, err)

	require.NoError(t, registry.Close())
	require.NoError(t, registry.Close())
}

func TestRegistryRegisterReturnsNotSupported(t *testing.T) {
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, &fakeDiscoveryClient{})
	require.NoError(t, err)

	err = registry.Register(context.Background(), "svc", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegisterNotSupported)
}

func TestRegistryUnderlyingReturnsClient(t *testing.T) {
	client := &fakeDiscoveryClient{}
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	require.Same(t, client, registry.Underlying())
}

func TestRegistryAsProjectsNativeClient(t *testing.T) {
	native := fake.NewSimpleClientset()
	client := &fakeNativeDiscoveryClient{native: native}
	registry, err := NewRegistryWithClient(&KubernetesConfig{Namespace: "default"}, client)
	require.NoError(t, err)

	var projected *fake.Clientset
	require.True(t, registry.As(&projected))
	require.Same(t, native, projected)
}

type fakeDiscoveryClient struct {
	discoverResult []contract.ServiceInstance
	discoverErr    error
	updateCh       chan []contract.ServiceInstance
	discoverCalls  int
}

type fakeNativeDiscoveryClient struct {
	fakeDiscoveryClient
	native any
}

func (f *fakeNativeDiscoveryClient) Underlying() any {
	return f.native
}

func (f *fakeDiscoveryClient) Discover(ctx context.Context, namespace, name string) ([]contract.ServiceInstance, error) {
	f.discoverCalls++
	if f.discoverErr != nil {
		return nil, f.discoverErr
	}
	return append([]contract.ServiceInstance(nil), f.discoverResult...), nil
}

func (f *fakeDiscoveryClient) Watch(ctx context.Context, namespace, name string, onUpdate func([]contract.ServiceInstance)) error {
	if f.updateCh == nil {
		f.updateCh = make(chan []contract.ServiceInstance, 2)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-f.updateCh:
			if errors.Is(f.discoverErr, ErrServiceNotFound) {
				continue
			}
			onUpdate(append([]contract.ServiceInstance(nil), update...))
		}
	}
}

func (f *fakeDiscoveryClient) push(update []contract.ServiceInstance) {
	if f.updateCh == nil {
		f.updateCh = make(chan []contract.ServiceInstance, 2)
	}
	f.updateCh <- append([]contract.ServiceInstance(nil), update...)
}
