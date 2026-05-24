package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
)

// ErrCacheMiss indicates that the cache key does not exist.
// ErrCacheMiss 表示缓存未命中。
var ErrCacheMiss = datacontract.ErrCacheMiss

// Cache is the top-level alias of the unified cache contract.
// Cache 是统一缓存契约的顶层别名。
type Cache = datacontract.Cache

// resolveCache 从框架容器解析缓存服务，内部辅助函数。
// 优先从 ctx 提取容器，其次回退到全局默认容器。
func resolveCache(ctx context.Context) (datacontract.Cache, error) {
	cont := frameworkcontainer.Resolve(ctx)
	if cont == nil {
		return nil, fmt.Errorf("cache: container not available from context or global default")
	}
	return container.MakeWith[datacontract.Cache](cont, datacontract.CacheKey)
}

// GetService 从容器获取统一缓存服务。
// 通过 ctx 解析框架容器后再解析缓存实例。
func GetService(ctx context.Context) (datacontract.Cache, error) {
	return resolveCache(ctx)
}

// MustGetService 从容器获取统一缓存服务，失败时 panic。
// 用于业务初始化阶段，确认缓存服务必须可用的场景。
func MustGetService(ctx context.Context) datacontract.Cache {
	svc, err := resolveCache(ctx)
	if err != nil {
		panic(err)
	}
	return svc
}

// Get 读取缓存。
// 根据 key 获取缓存值，若 key 不存在返回 ErrCacheMiss。
func Get(ctx context.Context, key string) (string, error) {
	cacheSvc, err := resolveCache(ctx)
	if err != nil {
		return "", err
	}
	return cacheSvc.Get(ctx, key)
}

// Set 写入缓存。
// 将 value 写入 key，并设置过期时间 ttl。
func Set(ctx context.Context, key, value string, ttl time.Duration) error {
	cacheSvc, err := resolveCache(ctx)
	if err != nil {
		return err
	}
	return cacheSvc.Set(ctx, key, value, ttl)
}

// Del 删除缓存。
// 删除指定 key 的缓存值。
func Del(ctx context.Context, key string) error {
	cacheSvc, err := resolveCache(ctx)
	if err != nil {
		return err
	}
	return cacheSvc.Del(ctx, key)
}

// MGet 批量读取缓存。
// 一次性获取多个 key 的值，返回 key→value 映射。
func MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	cacheSvc, err := resolveCache(ctx)
	if err != nil {
		return nil, err
	}
	return cacheSvc.MGet(ctx, keys...)
}

// Remember 读取缓存，未命中时回源计算并回填。
// 先尝试读取 key 的缓存值；若缓存未命中，则调用 fn 计算结果，
// 将结果写入缓存（ttl 过期时间）并返回。
//
// Example:
//
//	value, err := cache.Remember(ctx, "user:42", time.Minute, func(ctx context.Context) (string, error) {
//	    return loadUserJSON(ctx, 42)
//	})
func Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	cacheSvc, err := resolveCache(ctx)
	if err != nil {
		return "", err
	}
	return cacheSvc.Remember(ctx, key, ttl, fn)
}
