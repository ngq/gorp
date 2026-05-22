// Package cache provides Redis-based cache implementation.
// redisCache wraps datacontract.Redis to implement cacheDriver interface.
// Note: This implementation relies on Redis provider being registered first.
//
// 本文件提供基于 Redis 的缓存实现，适用于分布式系统或生产环境。
// redisCache 封装 datacontract.Redis 实现 cacheDriver 接口。
// 注意：此实现依赖 Redis Provider 先注册。
package cache

import (
	"context"
	"errors"
	"log/slog"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// redisCache implements cacheDriver using Redis backend.
//
// redisCache 使用 Redis 后端实现缓存驱动。
type redisCache struct {
	r datacontract.Redis // r is the Redis client.
	//
	// r Redis 客户端。
}

// newRedisCache creates a new Redis-based cache instance.
//
// newRedisCache 创建新的 Redis 缓存实例。
func newRedisCache(r datacontract.Redis) *redisCache {
	return &redisCache{r: r}
}

// Get retrieves a value by key. Returns ErrCacheMiss if key not found.
// Core logic: Call Redis Get, convert Redis nil error to ErrCacheMiss.
//
// Get 根据键获取值，未找到返回 ErrCacheMiss。
// 核心逻辑：调用 Redis Get，将 Redis nil 错误转换为 ErrCacheMiss。
func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.r.Get(ctx, key)
	if err != nil {
		if datacontract.IsRedisNil(err) {
			return "", datacontract.ErrCacheMiss
		}
		return "", err
	}
	return v, nil
}

// Set stores a key-value pair with TTL.
//
// Set 存储键值对并设置过期时间。
func (c *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.r.Set(ctx, key, value, ttl)
}

// Del deletes a key from Redis.
//
// Del 删除指定键。
func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.r.Del(ctx, key)
}

// MGet retrieves multiple keys at once using Redis MGET command.
//
// MGet 使用 Redis MGET 命令批量获取多个键的值。
func (c *redisCache) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return c.r.MGet(ctx, keys...)
}

// MSet writes multiple key-value pairs using Redis MSET command.
//
// MSet 使用 Redis MSET 命令批量写入多个键值对。
func (c *redisCache) MSet(ctx context.Context, kvs map[string]string, ttl time.Duration) error {
	if err := c.r.MSet(ctx, kvs); err != nil {
		return err
	}
	// Set TTL for each key if specified.
	// 为每个 key 设置 TTL（如果指定）。
	if ttl > 0 {
		for key := range kvs {
			if err := c.r.Expire(ctx, key, ttl); err != nil {
				return err
			}
		}
	}
	return nil
}

// redisBinaryCache implements BinaryCache using Redis backend.
//
// redisBinaryCache 使用 Redis 后端实现二进制安全缓存。
type redisBinaryCache struct {
	r datacontract.Redis
}

func newRedisBinaryCache(r datacontract.Redis) *redisBinaryCache {
	return &redisBinaryCache{r: r}
}

func (c *redisBinaryCache) Get(ctx context.Context, key string) ([]byte, error) {
	v, err := c.r.Get(ctx, key)
	if err != nil {
		if datacontract.IsRedisNil(err) {
			return nil, datacontract.ErrCacheMiss
		}
		return nil, err
	}
	return []byte(v), nil
}

func (c *redisBinaryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.r.Set(ctx, key, string(value), ttl)
}

func (c *redisBinaryCache) Del(ctx context.Context, key string) error {
	return c.r.Del(ctx, key)
}

func (c *redisBinaryCache) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	result, err := c.r.MGet(ctx, keys...)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]byte, len(result))
	for k, v := range result {
		out[k] = []byte(v)
	}
	return out, nil
}

func (c *redisBinaryCache) MSet(ctx context.Context, kvs map[string][]byte, ttl time.Duration) error {
	strKvs := make(map[string]string, len(kvs))
	for k, v := range kvs {
		strKvs[k] = string(v)
	}
	if err := c.r.MSet(ctx, strKvs); err != nil {
		return err
	}
	if ttl > 0 {
		for key := range kvs {
			if err := c.r.Expire(ctx, key, ttl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *redisBinaryCache) Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) ([]byte, error)) ([]byte, error) {
	v, err := c.Get(ctx, key)
	if err == nil {
		return v, nil
	}
	if !errors.Is(err, datacontract.ErrCacheMiss) {
		return nil, err
	}

	computed, err := fn(ctx)
	if err != nil {
		return nil, err
	}
	if setErr := c.Set(ctx, key, computed, ttl); setErr != nil {
		slog.Warn("cache: Remember failed to Set computed value", "key", key, "error", setErr)
	}
	return computed, nil
}
