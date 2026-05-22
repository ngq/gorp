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

// BinaryCacheKey is the container key for the binary cache capability.
//
// BinaryCacheKey 是二进制缓存能力的容器键。
const BinaryCacheKey = "framework.binary_cache"

// ErrCacheMiss indicates that the requested cache value does not exist.
//
// ErrCacheMiss 表示请求的缓存值不存在。
var ErrCacheMiss = errors.New("cache miss")

// Cache defines the common cache operations exposed by the framework.
// Uses string values for simplicity and JSON compatibility.
// For binary-safe caching, use BinaryCache which uses []byte values.
//
// Cache 定义框架对外暴露的通用缓存操作。
// 使用 string 值类型以简化使用和 JSON 兼容。
// 如需二进制安全缓存，请使用 BinaryCache（使用 []byte 值类型）。
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

	// MSet writes multiple key-value pairs in one call.
	//
	// MSet 一次调用写入多个键值对。
	MSet(ctx context.Context, kvs map[string]string, ttl time.Duration) error

	// Remember loads from cache first and computes the value on miss.
	//
	// Remember 先查缓存，未命中时再计算并回填结果。
	Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error)
}

// BinaryCache defines binary-safe cache operations using []byte values.
// Use this interface when caching binary data (protobuf, images, etc.).
//
// BinaryCache 定义使用 []byte 值类型的二进制安全缓存操作。
// 当缓存二进制数据（protobuf、图片等）时使用此接口。
type BinaryCache interface {
	// Get reads a cached []byte value by key.
	//
	// Get 按 key 读取缓存 []byte 值。
	Get(ctx context.Context, key string) ([]byte, error)

	// Set writes a []byte value into cache with a TTL.
	//
	// Set 按 TTL 将 []byte 值写入缓存。
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Del deletes a cached value by key.
	//
	// Del 按 key 删除缓存值。
	Del(ctx context.Context, key string) error

	// MGet reads multiple cached values in one call.
	//
	// MGet 一次调用读取多个缓存值。
	MGet(ctx context.Context, keys ...string) (map[string][]byte, error)

	// MSet writes multiple key-value pairs in one call.
	//
	// MSet 一次调用写入多个键值对。
	MSet(ctx context.Context, kvs map[string][]byte, ttl time.Duration) error

	// Remember loads from cache first and computes the value on miss.
	//
	// Remember 先查缓存，未命中时再计算并回填结果。
	Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) ([]byte, error)) ([]byte, error)
}
