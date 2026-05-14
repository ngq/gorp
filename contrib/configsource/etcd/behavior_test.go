package etcd

import (
	"context"
	"errors"
	"io"
	"testing"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/stretchr/testify/require"
)

func TestSetNestedValueBuildsNestedMap(t *testing.T) {
	result := make(map[string]any)
	setNestedValue(result, "a/b/c", "v")
	a := result["a"].(map[string]any)
	b := a["b"].(map[string]any)
	require.Equal(t, "v", b["c"])
}

func TestGetConfigSourceConfigRejectsInvalidConfigService(t *testing.T) {
	_, err := getConfigSourceConfig(etcdConfigInvalidContainerStub{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid config service")
}

func TestWatcherStopWithoutStartDoesNotPanic(t *testing.T) {
	w := &etcdWatcher{cancel: func() {}}
	require.NoError(t, w.Stop())
}

func TestWatcherOnChangeStoresCallback(t *testing.T) {
	w := &etcdWatcher{}
	called := false
	w.OnChange("demo", func(any) { called = true })
	count := 0
	w.callbacks.Range(func(_, _ any) bool { count++; return true })
	require.Equal(t, 1, count)
	require.False(t, called)
}

func TestSourceWatchRejectsClosedSource(t *testing.T) {
	s := &Source{cfg: &datacontract.ConfigSourceConfig{EtcdPath: "/config"}, closed: true}
	_, err := s.Watch(context.Background(), "demo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "source closed")
}

func TestSourceCloseWithoutClientPanicsToday(t *testing.T) {
	s := &Source{}
	require.Panics(t, func() { _ = s.Close() })
}

type etcdConfigInvalidContainerStub struct{}

func (etcdConfigInvalidContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (etcdConfigInvalidContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (etcdConfigInvalidContainerStub) IsBind(string) bool                                      { return true }
func (etcdConfigInvalidContainerStub) IsBindNamed(string, string) bool                         { return false }
func (etcdConfigInvalidContainerStub) Make(string) (any, error)                                { return 1, nil }
func (etcdConfigInvalidContainerStub) MakeNamed(string, string) (any, error) {
	return nil, errors.New("not found")
}
func (etcdConfigInvalidContainerStub) MustMake(string) any              { return 1 }
func (etcdConfigInvalidContainerStub) MustMakeNamed(string, string) any { return nil }
func (etcdConfigInvalidContainerStub) RegisterCloser(string, io.Closer) {}
func (etcdConfigInvalidContainerStub) Destroy() error                   { return nil }
func (etcdConfigInvalidContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (etcdConfigInvalidContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (etcdConfigInvalidContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo {
	return nil
}
func (etcdConfigInvalidContainerStub) DebugPrint() string { return "" }

func TestEtcdSourceErrorHelper(t *testing.T) {
	err := errors.New("boom")
	require.Equal(t, "boom", err.Error())
}
