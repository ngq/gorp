package eureka

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestProviderContract(t *testing.T) {
	p := NewProvider()
	require.Equal(t, "registry.eureka", p.Name())
	require.True(t, p.IsDefer())
	require.Equal(t, []string{contract.RPCRegistryKey}, p.Provides())
}

func TestRegistryRegisterUsesClient(t *testing.T) {
	client := &fakeEurekaClient{}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)
	require.Equal(t, 1, client.registerCalls)
}

func TestRegistryRegisterRejectsDuplicateInstance(t *testing.T) {
	client := &fakeEurekaClient{}
	cfg := testEurekaConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrAlreadyRegistered)
}

func TestRegistryDeregisterUsesClient(t *testing.T) {
	client := &fakeEurekaClient{}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.NoError(t, err)
	require.Equal(t, 1, client.deregisterCalls)
}

func TestRegistryDiscoverUsesClient(t *testing.T) {
	client := &fakeEurekaClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "USER-SERVICE:10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
	}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	instances, err := registry.Discover(context.Background(), "user-service")
	require.NoError(t, err)
	require.Len(t, instances, 1)
	require.Equal(t, "10.0.0.1:8080", instances[0].Address)
}

func TestRegistryDiscoverReturnsNotFound(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{
		discoverErr: ErrServiceNotFound,
	})
	require.NoError(t, err)

	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrServiceNotFound)
}

func TestRegistryRegisterReturnsSourceError(t *testing.T) {
	expected := errors.New("eureka: register failed")
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{
		registerErr: expected,
	})
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, expected)
}

func TestRegistryUnderlyingReturnsClient(t *testing.T) {
	client := &fakeEurekaClient{}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	require.Same(t, client, registry.Underlying())
}

func TestRegistryAsProjectsClient(t *testing.T) {
	client := &fakeEurekaClient{}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	var projected *fakeEurekaClient
	require.True(t, registry.As(&projected))
	require.Same(t, client, projected)
}

func TestRegistryAsRejectsInvalidTarget(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{})
	require.NoError(t, err)

	require.False(t, registry.As(nil))
	require.False(t, registry.As(fakeEurekaClient{}))
}

func TestRegistryHTTPClientProviderAvailableOnDefaultClient(t *testing.T) {
	registry, err := NewRegistry(testEurekaConfig())
	require.NoError(t, err)

	var provider HTTPClientProvider
	require.True(t, registry.As(&provider))
	require.NotNil(t, provider.HTTPClient())
}

func TestRegistryDeregisterReturnsSourceError(t *testing.T) {
	expected := errors.New("eureka: deregister failed")
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{
		deregisterErr: expected,
	})
	require.NoError(t, err)

	err = registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080")
	require.ErrorIs(t, err, expected)
}

func TestNewRegistryRequiresServerURL(t *testing.T) {
	registry, err := NewRegistryWithClient(&EurekaConfig{}, &fakeEurekaClient{})
	require.Nil(t, registry)
	require.ErrorIs(t, err, ErrNoServerURL)
}

func TestNewRegistryRequiresClient(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), nil)
	require.Nil(t, registry)
	require.EqualError(t, err, "eureka: client is required")
}

func TestRegistryCloseRejectsOperations(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.ErrorIs(t, err, ErrRegistryClosed)

	_, err = registry.Discover(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchDeliversInitialAndUpdatedSnapshot(t *testing.T) {
	client := &fakeEurekaClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "USER-SERVICE:10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchUpdates: make(chan []contract.ServiceInstance, 1),
	}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
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

	client.push([]contract.ServiceInstance{
		{ID: "USER-SERVICE:10.0.0.2:8080", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.2:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated watch snapshot")
	}
}

func TestRegistryWatchSkipsDuplicateSnapshot(t *testing.T) {
	client := &fakeEurekaClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "USER-SERVICE:10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchUpdates: make(chan []contract.ServiceInstance, 2),
	}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	client.push([]contract.ServiceInstance{
		{ID: "USER-SERVICE:10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		t.Fatalf("expected duplicate snapshot to be suppressed, got %#v", instances)
	case <-time.After(150 * time.Millisecond):
	}
}

func TestRegistryWatchAfterCloseFails(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())

	_, err = registry.Watch(context.Background(), "user-service")
	require.ErrorIs(t, err, ErrRegistryClosed)
}

func TestRegistryWatchChannelClosesAfterRegistryClose(t *testing.T) {
	client := &fakeEurekaClient{}
	registry, err := NewRegistryWithClient(testEurekaConfig(), client)
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

func TestRegistryCloseIsIdempotent(t *testing.T) {
	registry, err := NewRegistryWithClient(testEurekaConfig(), &fakeEurekaClient{})
	require.NoError(t, err)
	require.NoError(t, registry.Close())
	require.NoError(t, registry.Close())
}

func TestRegistryHeartbeatRunsAfterRegister(t *testing.T) {
	client := &fakeEurekaClient{}
	cfg := testEurekaConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.heartbeatCalls > 0
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryDeregisterStopsHeartbeat(t *testing.T) {
	client := &fakeEurekaClient{}
	cfg := testEurekaConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", nil)
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return client.heartbeatCalls > 0
	}, time.Second, 10*time.Millisecond)

	before := client.heartbeatCalls
	require.NoError(t, registry.Deregister(context.Background(), "user-service", "10.0.0.1:8080"))
	time.Sleep(40 * time.Millisecond)
	require.Equal(t, before, client.heartbeatCalls)
}

func TestRegistryHeartbeatNotFoundTriggersReRegister(t *testing.T) {
	client := &fakeEurekaClient{
		heartbeatErrs: []error{ErrServiceNotFound, nil},
	}
	cfg := testEurekaConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatRetryBackoff = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.registerCalls >= 2
	}, time.Second, 10*time.Millisecond)
}

func TestRegistryHeartbeatSourceErrorRetriesWithoutReRegister(t *testing.T) {
	client := &fakeEurekaClient{
		heartbeatErr: errors.New("eureka unavailable"),
	}
	cfg := testEurekaConfig()
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatRetryBackoff = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	err = registry.Register(context.Background(), "user-service", "10.0.0.1:8080", map[string]string{"version": "v1"})
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return client.heartbeatCalls >= 2
	}, time.Second, 10*time.Millisecond)
	require.Equal(t, 1, client.registerCalls)
}

func TestHTTPEurekaClientWatchRetriesAfterSourceError(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/eureka/apps/USER-SERVICE" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		requests++
		switch requests {
		case 1:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"application":{"instance":[{"instanceId":"USER-SERVICE:10.0.0.1:8080","app":"USER-SERVICE","ipAddr":"10.0.0.1","status":"UP","metadata":{},"port":{"$":8080}}]}}`))
		case 2:
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"application":{"instance":[{"instanceId":"USER-SERVICE:10.0.0.2:8080","app":"USER-SERVICE","ipAddr":"10.0.0.2","status":"UP","metadata":{},"port":{"$":8080}}]}}`))
		}
	}))
	defer server.Close()

	client := newHTTPEurekaClient().(*httpEurekaClient)
	cfg := testEurekaConfig()
	cfg.ServerURL = server.URL
	cfg.WatchInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updates := make(chan []contract.ServiceInstance, 4)
	go func() {
		_ = client.Watch(ctx, cfg, "user-service", func(instances []contract.ServiceInstance) {
			updates <- instances
		})
	}()

	select {
	case instances := <-updates:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.1:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected initial watch snapshot")
	}

	select {
	case instances := <-updates:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.2:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated watch snapshot after retry")
	}
}

func TestRegistryWatchRetriesAfterClientWatchError(t *testing.T) {
	client := &fakeEurekaClient{
		discoverResult: []contract.ServiceInstance{
			{ID: "USER-SERVICE:10.0.0.1:8080", Name: "user-service", Address: "10.0.0.1:8080", Healthy: true},
		},
		watchUpdates: make(chan []contract.ServiceInstance, 1),
		watchErrs:    []error{errors.New("temporary watch error")},
	}
	cfg := testEurekaConfig()
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
		return client.watchCalls >= 2
	}, time.Second, 10*time.Millisecond)

	client.push([]contract.ServiceInstance{
		{ID: "USER-SERVICE:10.0.0.2:8080", Name: "user-service", Address: "10.0.0.2:8080", Healthy: true},
	})

	select {
	case instances := <-ch:
		require.Len(t, instances, 1)
		require.Equal(t, "10.0.0.2:8080", instances[0].Address)
	case <-time.After(2 * time.Second):
		t.Fatal("expected updated watch snapshot after watch retry")
	}
}

func TestRegistryWatchStopsOnNonRetryableError(t *testing.T) {
	client := &fakeEurekaClient{
		watchErrs: []error{context.Canceled},
	}
	cfg := testEurekaConfig()
	cfg.WatchInterval = 10 * time.Millisecond
	registry, err := NewRegistryWithClient(cfg, client)
	require.NoError(t, err)

	ch, err := registry.Watch(context.Background(), "user-service")
	require.NoError(t, err)

	select {
	case _, ok := <-ch:
		if ok {
			select {
			case _, ok = <-ch:
				require.False(t, ok)
			case <-time.After(2 * time.Second):
				t.Fatal("expected watch channel to close after initial snapshot")
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("expected watch channel to close")
	}
	require.Equal(t, 1, client.watchCalls)
}

type fakeEurekaClient struct {
	registerErr     error
	deregisterErr   error
	heartbeatErr    error
	heartbeatErrs   []error
	discoverErr     error
	discoverResult  []contract.ServiceInstance
	discoverErrs    []error
	discoverResults [][]contract.ServiceInstance
	watchUpdates    chan []contract.ServiceInstance
	watchErrs       []error
	registerCalls   int
	deregisterCalls int
	heartbeatCalls  int
	discoverCalls   int
	watchCalls      int
}

func (f *fakeEurekaClient) Register(ctx context.Context, cfg *EurekaConfig, name, addr string, meta map[string]string) error {
	f.registerCalls++
	return f.registerErr
}

func (f *fakeEurekaClient) Deregister(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	f.deregisterCalls++
	return f.deregisterErr
}

func (f *fakeEurekaClient) Heartbeat(ctx context.Context, cfg *EurekaConfig, name, addr string) error {
	f.heartbeatCalls++
	if len(f.heartbeatErrs) > 0 {
		err := f.heartbeatErrs[0]
		f.heartbeatErrs = f.heartbeatErrs[1:]
		return err
	}
	return f.heartbeatErr
}

func (f *fakeEurekaClient) Discover(ctx context.Context, cfg *EurekaConfig, name string) ([]contract.ServiceInstance, error) {
	f.discoverCalls++
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
		return append([]contract.ServiceInstance(nil), result...), nil
	}
	return append([]contract.ServiceInstance(nil), f.discoverResult...), nil
}

func (f *fakeEurekaClient) Watch(ctx context.Context, cfg *EurekaConfig, name string, onUpdate func([]contract.ServiceInstance)) error {
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

func (f *fakeEurekaClient) push(instances []contract.ServiceInstance) {
	if f.watchUpdates == nil {
		f.watchUpdates = make(chan []contract.ServiceInstance, 1)
	}
	f.watchUpdates <- instances
}

func testEurekaConfig() *EurekaConfig {
	return &EurekaConfig{
		ServerURL:             "http://eureka.local",
		AppName:               "demo",
		HeartbeatRetryBackoff: time.Second,
	}
}
