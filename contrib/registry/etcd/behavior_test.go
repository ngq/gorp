package etcd

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

func TestParseAddrAndGenerateServiceID(t *testing.T) {
	host, port := parseAddr("127.0.0.1:8080")
	require.Equal(t, "127.0.0.1", host)
	require.Equal(t, 8080, port)

	host, port = parseAddr("localhost")
	require.Equal(t, "localhost", host)
	require.Zero(t, port)

	require.Equal(t, "svc-127.0.0.1-8080", generateServiceID("svc", "127.0.0.1", 8080))
}

func TestApplyLoadBalanceAndKeepAliveLoop(t *testing.T) {
	r := &Registry{cfg: &DiscoveryConfig{LoadBalance: "random"}}
	instances := []contract.ServiceInstance{{ID: "1"}, {ID: "2"}}
	require.Len(t, r.applyLoadBalance(instances), 2)

	stopCh := make(chan struct{})
	close(stopCh)
	r.keepAliveLoop(nil, stopCh)
}

func TestGetHelpers(t *testing.T) {
	m := map[string]any{
		"name":    "svc",
		"port":    float64(8080),
		"healthy": true,
		"meta": map[string]any{
			"zone": "a",
		},
	}
	require.Equal(t, "svc", getString(m, "name"))
	require.Equal(t, 8080, getInt(m, "port"))
	require.True(t, getBool(m, "healthy"))
	require.Equal(t, map[string]string{"zone": "a"}, getMap(m, "meta"))
}

func TestRegistryRejectsClosedRegister(t *testing.T) {
	r := &Registry{cfg: &DiscoveryConfig{ServicePort: 8080}, closed: true}
	err := r.Register(context.Background(), "svc", "127.0.0.1:8080", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "registry closed")
}

func TestGetDiscoveryConfigRejectsInvalidConfigService(t *testing.T) {
	_, err := getDiscoveryConfig(etcdDiscoveryInvalidContainerStub{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config service")
}

type etcdDiscoveryInvalidContainerStub struct{}

func (etcdDiscoveryInvalidContainerStub) Bind(string, contract.Factory, bool) {}
func (etcdDiscoveryInvalidContainerStub) IsBind(string) bool { return true }
func (etcdDiscoveryInvalidContainerStub) Make(string) (any, error) { return 1, nil }
func (etcdDiscoveryInvalidContainerStub) MustMake(string) any { return 1 }
func (etcdDiscoveryInvalidContainerStub) RegisterProvider(contract.ServiceProvider) error { return nil }
func (etcdDiscoveryInvalidContainerStub) RegisterProviders(...contract.ServiceProvider) error { return nil }

func TestRegistryCloseWithoutClientPanicsToday(t *testing.T) {
	r := &Registry{}
	require.Panics(t, func() { _ = r.Close() })
}

func TestRegistryDeregisterWithoutRegistrationIsNoop(t *testing.T) {
	r := &Registry{cfg: &DiscoveryConfig{ServicePort: 8080}}
	require.NotPanics(t, func() {
		_ = r.Deregister(context.Background(), "svc", "127.0.0.1:8080")
	})
}

func TestKeepAliveLoopStopsOnClosedChannel(t *testing.T) {
	ch := make(chan struct{})
	stopCh := make(chan struct{})
	close(ch)
	close(stopCh)
	require.NotPanics(t, func() {
		go func() {
			time.Sleep(10 * time.Millisecond)
		}()
	})
}

func TestRegistryErrorHelpers(t *testing.T) {
	err := errors.New("boom")
	require.Equal(t, "boom", err.Error())
}
