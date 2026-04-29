package cache

import (
	"context"
	"testing"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/stretchr/testify/require"
)

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
func (s *exportCacheStub) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	return fn(ctx)
}

type exportCacheContainerStub struct {
	cache contract.Cache
}

func (s *exportCacheContainerStub) Bind(string, contract.Factory, bool)                  {}
func (s *exportCacheContainerStub) IsBind(string) bool                                   { return true }
func (s *exportCacheContainerStub) MustMake(key string) any                              { v, _ := s.Make(key); return v }
func (s *exportCacheContainerStub) RegisterProvider(contract.ServiceProvider) error       { return nil }
func (s *exportCacheContainerStub) RegisterProviders(...contract.ServiceProvider) error   { return nil }
func (s *exportCacheContainerStub) Make(key string) (any, error) {
	if key == contract.CacheKey {
		return s.cache, nil
	}
	return nil, context.DeadlineExceeded
}

func TestExportedCacheHelpers(t *testing.T) {
	stub := &exportCacheStub{value: "v1"}
	containerStub := &exportCacheContainerStub{cache: stub}

	cacheSvc, err := Make(containerStub)
	require.NoError(t, err)
	require.Same(t, stub, cacheSvc)
	require.Same(t, stub, MustMake(containerStub))

	value, err := Get(context.Background(), containerStub, "k")
	require.NoError(t, err)
	require.Equal(t, "v1", value)

	err = Set(context.Background(), containerStub, "user:1", "alice", time.Minute)
	require.NoError(t, err)
	require.Equal(t, "user:1", stub.key)
	require.Equal(t, "alice", stub.value)
	require.Equal(t, time.Minute, stub.ttl)

	result, err := Remember(context.Background(), containerStub, "k", time.Second, func(context.Context) (string, error) {
		return "computed", nil
	})
	require.NoError(t, err)
	require.Equal(t, "computed", result)
	require.Same(t, contract.ErrCacheMiss, ErrCacheMiss)
}
