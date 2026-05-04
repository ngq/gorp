package data

import (
	"context"
	"errors"
	"time"
)

const CacheKey = "framework.cache"

var ErrCacheMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	MGet(ctx context.Context, keys ...string) (map[string]string, error)
	Remember(ctx context.Context, key string, ttl time.Duration, fn func(context.Context) (string, error)) (string, error)
}
