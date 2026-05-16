// Application scenarios:
// - Define the cache contract used by business services and framework providers.
// - Standardize cache read, write, delete, and remember-style loading semantics.
// - Expose shared cache constants and sentinel errors across implementations.
//
// 适用场景：
// - 定义业务服务和框架 provider 共同使用的缓存契约。
// - 统一缓存读取、写入、删除和 remember 风格加载语义。
// - 在不同实现之间共享缓存常量与哨兵错误。
package data

import (
	"context"
	"errors"
	"time"
)

// CacheKey is the container key for the cache capability.
//
// CacheKey 是缓存能力的容器键。
const CacheKey = "framework.cache"

// ErrCacheMiss indicates that the requested cache value does not exist.
//
// ErrCacheMiss 表示请求的缓存值不存在。
var ErrCacheMiss = errors.New("cache miss")

// Cache defines the common cache operations exposed by the framework.
//
// Future improvements:
//   - Consider changing value type from string to []byte for binary-safe caching.
//   - Consider adding MSet for batch write symmetry with MGet.
//
// Cache 定义框架对外暴露的通用缓存操作。
//
// 未来改进：
//   - 考虑将值类型从 string 改为 []byte 以支持二进制安全缓存。
//   - 考虑添加 MSet 以与 MGet 批量读取对称。
type Cache interface {
	// Get reads a cached string value by key.
	//
	// Get 按 key 读取缓存字符串值。
	Get(ctx context.Context, key string) (string, error)

	// Set writes a string value into cache with a TTL.
	//
	// Set 按 TTL 将字符串值写入缓存。
	Set(ctx context.Context, key, value string, ttl time.Duration) error

	// Del deletes a cached value by key.
	//
	// Del 按 key 删除缓存值。
	Del(ctx context.Context, key string) error

	// MGet reads multiple cached values in one call.
	//
	// MGet 一次调用读取多个缓存值。
	MGet(ctx context.Context, keys ...string) (map[string]string, error)

	// Remember loads from cache first and computes the value on miss.
	//
	// Remember 先查缓存，未命中时再计算并回填结果。
	Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error)
}
