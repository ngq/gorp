package cache

import (
	"context"
	"io"
	"testing"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/stretchr/testify/require"
)

// exportCacheStub 实现缓存契约，用于测试验证 facade 函数转发是否正确。
type exportCacheStub struct {
	value string
	key   string
	ttl   time.Duration
}

func (s *exportCacheStub) Get(context.Context, string) (string, error) { return s.value, nil }
func (s *exportCacheStub) Set(_ context.Context, key, value string, ttl time.Duration) error {
	s.key = key
	s.value = value
	s.ttl = ttl
	return nil
}
func (s *exportCacheStub) Del(context.Context, string) error { return nil }
func (s *exportCacheStub) MGet(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"k": s.value}, nil
}
func (s *exportCacheStub) MSet(context.Context, map[string]string, time.Duration) error {
	return nil
}
func (s *exportCacheStub) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	return fn(ctx)
}

// exportContainerStub 最小化容器桩，仅实现 Make 以返回缓存桩实例。
type exportContainerStub struct {
	cache datacontract.Cache
}

func (s *exportContainerStub) Bind(string, runtimecontract.Factory, bool)              {}
func (s *exportContainerStub) NamedBind(string, string, runtimecontract.Factory, bool) {}
func (s *exportContainerStub) IsBind(string) bool                                      { return true }
func (s *exportContainerStub) IsBindNamed(string, string) bool                         { return false }
func (s *exportContainerStub) MustMake(key string) any                                 { v, _ := s.Make(key); return v }
func (s *exportContainerStub) MustMakeNamed(string, string) any                        { return nil }
func (s *exportContainerStub) Make(key string) (any, error) {
	if key == datacontract.CacheKey {
		return s.cache, nil
	}
	return nil, context.DeadlineExceeded
}
func (s *exportContainerStub) MakeNamed(string, string) (any, error) { return nil, nil }
func (s *exportContainerStub) RegisterProvider(runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportContainerStub) RegisterProviders(...runtimecontract.ServiceProvider) error {
	return nil
}
func (s *exportContainerStub) RegisterCloser(string, io.Closer)                    {}
func (s *exportContainerStub) Destroy() error                                      { return nil }
func (s *exportContainerStub) RegisteredProviders() []runtimecontract.ProviderInfo { return nil }
func (s *exportContainerStub) DebugPrint() string                                  { return "" }
func (s *exportContainerStub) ProviderDAG() runtimecontract.ProviderDAG {
	return runtimecontract.ProviderDAG{}
}

// setupContainer 注入桩容器到全局默认，测试结束后自动清理。
func setupContainer(cache datacontract.Cache) {
	frameworkcontainer.SetDefault(&exportContainerStub{cache: cache})
}

func TestExportedCacheHelpers(t *testing.T) {
	stub := &exportCacheStub{value: "v1"}
	setupContainer(stub)
	t.Cleanup(func() { frameworkcontainer.SetDefault(nil) })

	ctx := context.Background()

	// GetService / MustGetService
	cacheSvc, err := GetService(ctx)
	require.NoError(t, err)
	require.Same(t, stub, cacheSvc)
	require.Same(t, stub, MustGetService(ctx))

	// Get
	value, err := Get(ctx, "k")
	require.NoError(t, err)
	require.Equal(t, "v1", value)

	// Set
	err = Set(ctx, "user:1", "alice", time.Minute)
	require.NoError(t, err)
	require.Equal(t, "user:1", stub.key)
	require.Equal(t, "alice", stub.value)
	require.Equal(t, time.Minute, stub.ttl)

	// Remember
	result, err := Remember(ctx, "k", time.Second, func(context.Context) (string, error) {
		return "computed", nil
	})
	require.NoError(t, err)
	require.Equal(t, "computed", result)

	// ErrCacheMiss 别名
	require.Same(t, datacontract.ErrCacheMiss, ErrCacheMiss)
}
