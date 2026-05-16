// Application scenarios:
// - Define the Redis contract used by providers and business helpers.
// - Standardize basic Redis key-value semantics without exposing a concrete client type.
// - Share Redis-related constants across integrations and middleware.
//
// 适用场景：
// - 定义 provider 和业务辅助代码共同使用的 Redis 契约。
// - 在不暴露具体客户端类型的前提下统一基础 Redis 键值语义。
// - 在集成和中间件之间共享 Redis 相关常量。
package data

import (
	"context"
	"strings"
	"time"
)

// RedisKey is the container key for the Redis capability.
//
// RedisKey 是 Redis 能力的容器键。
const RedisKey = "framework.redis"

// RedisNilString is the canonical missing-value string returned by Redis clients.
//
// RedisNilString 是 Redis 客户端返回空值时的标准字符串表示。
const RedisNilString = "redis: nil"

// IsRedisNil checks whether the given error represents a Redis nil (key not found) response.
// This is more robust than raw string matching and should be used instead of
// strings.Contains(err.Error(), RedisNilString).
//
// IsRedisNil 检查给定错误是否为 Redis nil（键不存在）响应。
// 比直接字符串匹配更健壮，应优先使用。
func IsRedisNil(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), RedisNilString)
}

// Redis defines the minimal Redis operations exposed by the framework.
//
// Redis 定义框架对外暴露的最小 Redis 操作集合。
type Redis interface {
	// Ping checks whether the Redis connection is available.
	//
	// Ping 检查 Redis 连接是否可用。
	Ping(ctx context.Context) error

	// Get reads a string value by key.
	//
	// Get 按 key 读取字符串值。
	Get(ctx context.Context, key string) (string, error)

	// Set writes a string value with a TTL.
	//
	// Set 按 TTL 写入字符串值。
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Del deletes a key.
	//
	// Del 删除指定 key。
	Del(ctx context.Context, key string) error

	// MGet reads multiple string values in one call.
	//
	// MGet 一次调用读取多个字符串值。
	MGet(ctx context.Context, keys ...string) (map[string]string, error)

	// MSet writes multiple string values in one call.
	//
	// MSet 一次调用写入多个字符串值。
	MSet(ctx context.Context, kvs map[string]string) error

	// Expire sets a TTL on an existing key.
	//
	// Expire 为已存在的 key 设置 TTL。
	Expire(ctx context.Context, key string, ttl time.Duration) error
}
