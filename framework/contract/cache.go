package contract

import (
	"context"
	"errors"
	"time"
)

const CacheKey = "framework.cache"

var ErrCacheMiss = errors.New("cache miss")

// Cache is a cache-aside oriented cache service.
//
// Conventions:
// - On key miss, return ErrCacheMiss.
// - ttl <= 0 means no expiration.
//
// 中文说明：
// - 当前支持内存驱动和 Redis 驱动。
// - 提供 TTL、Del、Remember 等核心能力。
// - 后续可扩展分布式缓存、多级缓存等高级特性。
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	MGet(ctx context.Context, keys ...string) (map[string]string, error)

	// Remember implements cache-aside:
	// - if key exists: return cached value
	// - if miss: call fn to compute, best-effort Set, then return
	Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error)
}
