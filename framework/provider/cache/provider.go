// Package cache provides a unified cache service abstraction for gorp framework.
// The package implements a cache service that can be configured via "cache.driver"
// config key or CACHE_DRIVER environment variable. Supported drivers:
// - "redis" (default): Redis-based distributed cache
// - "memory"/"mem"/"inmemory": In-memory local cache
// Eg:
//
// 缓存服务包，提供统一的缓存抽象，支持 Redis 和内存两种驱动。
// Eg:
//	// 在 bootstrap 中注册缓存 Provider
//	app.Register(cache.NewProvider())
//
//	// 通过配置指定驱动
//	// config.yaml:
//	// cache:
//	//   driver: redis
//
//	// 获取缓存服务
//	cacheSvc := c.MustMake(datacontract.CacheKey).(datacontract.Cache)
//	cacheSvc.Set(ctx, "user:123", "value", 10*time.Minute)
package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the unified cache contract.
//
// Provider 注册统一缓存契约，支持 Redis 和内存两种驱动。
type Provider struct{}

// NewProvider creates a new cache provider instance.
//
// NewProvider 创建新的缓存 Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "cache".
//
// Name 返回 Provider 名称 "cache"。
func (p *Provider) Name() string { return "cache" }

// IsDefer returns false, cache should be initialized immediately.
//
// IsDefer 返回 false，缓存服务应立即初始化。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the cache contract key.
//
// Provides 返回缓存契约键。
func (p *Provider) Provides() []string { return []string{datacontract.CacheKey} }

// config holds the cache configuration.
//
// config 缓存配置结构。
type config struct {
	Driver string `mapstructure:"driver"`
}

// Register binds the cache service factory to the container.
// It reads "cache.driver" from config or CACHE_DRIVER env, defaults to "redis".
//
// Register 将缓存服务工厂绑定到容器。
// 从配置读取 "cache.driver" 或环境变量 CACHE_DRIVER，默认为 "redis"。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.CacheKey, func(c runtimecontract.Container) (any, error) {
		cfgAny, err := c.Make(datacontract.ConfigKey)
		if err != nil {
			return nil, err
		}
		cfg := cfgAny.(datacontract.Config)

		var cc config
		_ = cfg.Unmarshal("cache", &cc)
		driver := strings.TrimSpace(cc.Driver)
		if driver == "" {
			driver = strings.TrimSpace(os.Getenv("CACHE_DRIVER"))
		}
		if driver == "" {
			driver = "redis"
		}

		var d cacheDriver
		switch strings.ToLower(driver) {
		case "memory", "mem", "inmemory":
			d = newMemoryStore()
		case "redis":
			rAny, err := c.Make(datacontract.RedisKey)
			if err != nil {
				return nil, err
			}
			r := rAny.(datacontract.Redis)
			d = newRedisCache(r)
		default:
			return nil, fmt.Errorf("invalid cache.driver: %s", driver)
		}

		return &service{d: d}, nil
	}, true)
	return nil
}

// Boot is a no-op for cache provider.
//
// Boot 缓存 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// cacheDriver defines the interface for cache backend implementations.
//
// cacheDriver 定义缓存后端接口，支持 Redis 和内存两种实现。
type cacheDriver interface {
	// Get retrieves a value by key. Returns ErrCacheMiss if not found.
	//
	// Get 根据键获取值，未找到返回 ErrCacheMiss。
	Get(ctx context.Context, key string) (string, error)
	// Set stores a key-value pair with TTL.
	//
	// Set 存储键值对并设置过期时间。
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	// Del deletes a key.
	//
	// Del 删除指定键。
	Del(ctx context.Context, key string) error
	// MGet retrieves multiple keys at once.
	//
	// MGet 批量获取多个键的值。
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
}

// service wraps cacheDriver to implement datacontract.Cache.
//
// service 封装 cacheDriver 实现 datacontract.Cache 接口。
type service struct {
	d cacheDriver
}

// Get retrieves a value by key.
//
// Get 根据键获取值。
func (s *service) Get(ctx context.Context, key string) (string, error) {
	return s.d.Get(ctx, key)
}

// Set stores a key-value pair with TTL.
//
// Set 存储键值对并设置过期时间。
func (s *service) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.d.Set(ctx, key, value, ttl)
}

// Del deletes a key.
//
// Del 删除指定键。
func (s *service) Del(ctx context.Context, key string) error {
	return s.d.Del(ctx, key)
}

// MGet retrieves multiple keys at once.
//
// MGet 批量获取多个键的值。
func (s *service) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return s.d.MGet(ctx, keys...)
}

// Remember implements the "cache-aside" pattern: get from cache, or compute and cache.
// Eg:
//
// Remember 实现 "cache-aside" 模式：先从缓存获取，不存在则计算后存入缓存。
// Eg:
//
//	val, err := cacheSvc.Remember(ctx, "user:profile:123", 10*time.Minute, func(ctx context.Context) (string, error) {
//	    return fetchUserProfileFromDB(ctx, 123)
//	})
func (s *service) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error) {
	v, err := s.d.Get(ctx, key)
	if err == nil {
		return v, nil
	}
	if err != datacontract.ErrCacheMiss {
		return "", err
	}

	computed, err := fn(ctx)
	if err != nil {
		return "", err
	}
	_ = s.d.Set(ctx, key, computed, ttl)
	return computed, nil
}