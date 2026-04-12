package contract

import "context"

const RedisKey = "framework.redis"

// RedisNilString is a sentinel error string used by go-redis for missing keys.
// We keep it here to avoid importing go-redis in higher layers.
const RedisNilString = "redis: nil"

type Redis interface {
	Ping(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttlSeconds int) error
	Del(ctx context.Context, key string) error
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
}
