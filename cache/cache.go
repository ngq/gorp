package cache

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// ErrCacheMiss indicates that the cache key does not exist.
// ErrCacheMiss 表示缓存未命中。
var ErrCacheMiss = datacontract.ErrCacheMiss

// Cache is the top-level alias of the unified cache contract.
// Cache 是统一缓存契约的顶层别名。
type Cache = datacontract.Cache

// Make returns the unified cache service from the container.
// Make 从容器获取统一缓存服务。
func Make(c runtimecontract.Container) (datacontract.Cache, error) {
	return container.MakeCache(c)
}

// MustMake returns the unified cache service from the container and panics on failure.
// MustMake 从容器获取统一缓存服务，失败 panic。
func MustMake(c runtimecontract.Container) datacontract.Cache {
	return container.MustMakeCache(c)
}

// Get reads a cache value by key.
// Get 读取缓存。
func Get(ctx context.Context, c runtimecontract.Container, key string) (string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return "", err
	}
	return cacheSvc.Get(ctx, key)
}

// Set writes a cache value with ttl.
// Set 写入缓存。
func Set(ctx context.Context, c runtimecontract.Container, key, value string, ttl time.Duration) error {
	cacheSvc, err := Make(c)
	if err != nil {
		return err
	}
	return cacheSvc.Set(ctx, key, value, ttl)
}

// Del deletes a cache key.
// Del 删除缓存。
func Del(ctx context.Context, c runtimecontract.Container, key string) error {
	cacheSvc, err := Make(c)
	if err != nil {
		return err
	}
	return cacheSvc.Del(ctx, key)
}

// MGet reads multiple cache keys in one call.
// MGet 批量读取缓存。
func MGet(ctx context.Context, c runtimecontract.Container, keys ...string) (map[string]string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return nil, err
	}
	return cacheSvc.MGet(ctx, keys...)
}

// Remember reads a cache value and falls back to the callback on cache miss.
// Remember 读取缓存，未命中时回源计算并回填。
//
// Example:
//
//	value, err := cache.Remember(ctx, c, "user:42", time.Minute, func(ctx context.Context) (string, error) {
//	    return loadUserJSON(ctx, 42)
//	})
func Remember(ctx context.Context, c runtimecontract.Container, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return "", err
	}
	return cacheSvc.Remember(ctx, key, ttl, fn)
}
