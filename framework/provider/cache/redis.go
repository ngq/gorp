package cache

import (
	"context"
	"strings"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

type redisCache struct {
	r datacontract.Redis
}

func newRedisCache(r datacontract.Redis) *redisCache {
	return &redisCache{r: r}
}

func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.r.Get(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), datacontract.RedisNilString) {
			return "", datacontract.ErrCacheMiss
		}
		return "", err
	}
	return v, nil
}

func (c *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.r.Set(ctx, key, value, int(ttl.Seconds()))
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.r.Del(ctx, key)
}

func (c *redisCache) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return c.r.MGet(ctx, keys...)
}
