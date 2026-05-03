package etcd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.etcd", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.RPCRegistryKey}, p.Provides())
}

func TestRegistryRegisterAndDiscover(t *testing.T) {
	client := newFakeEtcdRegistryClient()
	registry := NewRegistryWithClient(testEtcdConfig(), client)

	err := registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "user-service")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
	require.Equal(t, "v1", instances[0].Metadata["version"])
}

func TestRegistryKeepAliveFailureTriggersReRegister(t *testing.T) {
	client := newFakeEtcdRegistryClient()
	registry := NewRegistryWithClient(testEtcdConfig(), client)

	err := registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)
	require.Equal(t, 1, client.grantCalls)

	client.closeCurrentKeepAlive()

	require.Eventually(t, func() bool {
		return client.grantCalls >= 2 && client.putCalls >= 2
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryDeregisterStopsReRegister(t *testing.T) {
	client := newFakeEtcdRegistryClient()
	registry := NewRegistryWithClient(testEtcdConfig(), client)

	err := registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	serviceID := generateServiceID("user-service", "10.0.0.1", 8080)
	cached, ok := registry.registered.Load(serviceID)
	require.True(t, ok)
	reg := cached.(*registeredService)
	keepAliveCh := client.keepAliveChannels[reg.leaseID]
	require.NotNil(t, keepAliveCh)

	require.NoError(t, registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080"))
	close(keepAliveCh)

	time.Sleep(50 * time.Millisecond)
	require.Equal(t, 1, client.grantCalls)
}

func TestRegistryCloseRejectsOperations(t *testing.T) {
	client := newFakeEtcdRegistryClient()
	registry := NewRegistryWithClient(testEtcdConfig(), client)
	require.NoError(t, registry.Close())

	err := registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegistryClosed)

	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryDiscoverReturnsSourceError(t *testing.T) {
	client := newFakeEtcdRegistryClient()
	client.getErr = errors.New("etcd unavailable")
	registry := NewRegistryWithClient(testEtcdConfig(), client)

	_, err := registry.Discover(context.Background(), "user-service")
	require.ErrorContains(t, err, "get services failed")
}

func TestRegistryUnderlyingAndAs(t *testing.T) {
	native := &clientv3.Client{}
	client := &fakeNativeEtcdRegistryClient{
		fakeEtcdRegistryClient: newFakeEtcdRegistryClient(),
		native:                 native,
	}
	registry := NewRegistryWithClient(testEtcdConfig(), client)

	require.Same(t, native, registry.Underlying())

	var projected *clientv3.Client
	require.True(t, registry.As(&projected))
	require.Same(t, native, projected)
}

type fakeEtcdRegistryClient struct {
	mu                sync.Mutex
	kv                map[string]string
	getErr            error
	grantErr          error
	putErr            error
	keepAliveErr      error
	revokeErr         error
	grantCalls        int
	putCalls          int
	revokeCalls       int
	nextLeaseID       clientv3.LeaseID
	currentLeaseID    clientv3.LeaseID
	keepAliveChannels map[clientv3.LeaseID]chan *clientv3.LeaseKeepAliveResponse
}

type fakeNativeEtcdRegistryClient struct {
	*fakeEtcdRegistryClient
	native any
}

func (f *fakeNativeEtcdRegistryClient) Underlying() any {
	return f.native
}

func newFakeEtcdRegistryClient() *fakeEtcdRegistryClient {
	return &fakeEtcdRegistryClient{
		kv:                make(map[string]string),
		nextLeaseID:       1,
		keepAliveChannels: make(map[clientv3.LeaseID]chan *clientv3.LeaseKeepAliveResponse),
	}
}

func (f *fakeEtcdRegistryClient) Grant(ctx context.Context, ttl int64) (clientv3.LeaseID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.grantErr != nil {
		return 0, f.grantErr
	}
	f.grantCalls++
	leaseID := f.nextLeaseID
	f.nextLeaseID++
	f.currentLeaseID = leaseID
	return leaseID, nil
}

func (f *fakeEtcdRegistryClient) Put(ctx context.Context, key, value string, leaseID clientv3.LeaseID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.putErr != nil {
		return f.putErr
	}
	f.putCalls++
	f.kv[key] = value
	return nil
}

func (f *fakeEtcdRegistryClient) KeepAlive(ctx context.Context, leaseID clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.keepAliveErr != nil {
		return nil, f.keepAliveErr
	}
	ch := make(chan *clientv3.LeaseKeepAliveResponse, 1)
	f.keepAliveChannels[leaseID] = ch
	return ch, nil
}

func (f *fakeEtcdRegistryClient) Revoke(ctx context.Context, leaseID clientv3.LeaseID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.revokeErr != nil {
		return f.revokeErr
	}
	f.revokeCalls++
	delete(f.keepAliveChannels, leaseID)
	return nil
}

func (f *fakeEtcdRegistryClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}

	resp := &clientv3.GetResponse{}
	for k, v := range f.kv {
		if stringsHasPrefix(k, key) {
			resp.Kvs = append(resp.Kvs, &mvccpb.KeyValue{
				Key:   []byte(k),
				Value: []byte(v),
			})
		}
	}
	return resp, nil
}

func (f *fakeEtcdRegistryClient) Close() error { return nil }

func (f *fakeEtcdRegistryClient) closeCurrentKeepAlive() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if ch, ok := f.keepAliveChannels[f.currentLeaseID]; ok {
		close(ch)
		delete(f.keepAliveChannels, f.currentLeaseID)
	}
}

func testEtcdConfig() *DiscoveryConfig {
	return &DiscoveryConfig{
		EtcdEndpoints: []string{"127.0.0.1:2379"},
		ServicePath:   "/services/",
		LeaseTTL:      10,
		ServicePort:   8080,
		LoadBalance:   "random",
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
