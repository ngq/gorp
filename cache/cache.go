package cache

import (
	"context"
	"time"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

var ErrCacheMiss = datacontract.ErrCacheMiss

type Cache = datacontract.Cache

// Make 从容器获取统一缓存服务。
func Make(c runtimecontract.Container) (datacontract.Cache, error) {
	return container.MakeCache(c)
}

// MustMake 从容器获取统一缓存服务，失败 panic。
func MustMake(c runtimecontract.Container) datacontract.Cache {
	return container.MustMakeCache(c)
}

// Get 读取缓存。
func Get(ctx context.Context, c runtimecontract.Container, key string) (string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return "", err
	}
	return cacheSvc.Get(ctx, key)
}

// Set 写入缓存。
func Set(ctx context.Context, c runtimecontract.Container, key, value string, ttl time.Duration) error {
	cacheSvc, err := Make(c)
	if err != nil {
		return err
	}
	return cacheSvc.Set(ctx, key, value, ttl)
}

// Del 删除缓存。
func Del(ctx context.Context, c runtimecontract.Container, key string) error {
	cacheSvc, err := Make(c)
	if err != nil {
		return err
	}
	return cacheSvc.Del(ctx, key)
}

// MGet 批量读取缓存。
func MGet(ctx context.Context, c runtimecontract.Container, keys ...string) (map[string]string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return nil, err
	}
	return cacheSvc.MGet(ctx, keys...)
}

// Remember 读取缓存，未命中时回源计算并回填。
func Remember(ctx context.Context, c runtimecontract.Container, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	cacheSvc, err := Make(c)
	if err != nil {
		return "", err
	}
	return cacheSvc.Remember(ctx, key, ttl, fn)
}
