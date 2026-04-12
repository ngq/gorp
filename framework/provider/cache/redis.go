package cache

import (
	"context"
	"strings"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// redisCache 是 cacheDriver 的 Redis 版适配器。
//
// 中文说明：
// - 它本身不直接依赖 go-redis，而是依赖 contract.Redis 抽象。
// - 这样可以把真正的 Redis 客户端创建与配置逻辑留在 redis provider 中统一管理。
// - cache 层只关注“缓存语义”，例如命中、未命中、TTL，而不关心底层连接细节。
type redisCache struct {
	r contract.Redis
}

func newRedisCache(r contract.Redis) *redisCache {
	return &redisCache{r: r}
}

// Get 从 Redis 读取单个 key，并把底层 nil 语义转换为统一的 ErrCacheMiss。
func (c *redisCache) Get(ctx context.Context, key string) (string, error) {
	v, err := c.r.Get(ctx, key)
	if err != nil {
		// Distinguish key-miss from real error.
		//
		// 中文说明：
		// - Redis 在 key 不存在时通常返回特定 nil 错误，而不是空字符串。
		// - cache 抽象层不希望业务依赖底层驱动错误文本，因此这里统一翻译成 ErrCacheMiss。
		// - 真正的网络错误、权限错误等则原样返回，方便上层感知故障。
		if strings.Contains(err.Error(), contract.RedisNilString) {
			return "", contract.ErrCacheMiss
		}
		return "", err
	}
	return v, nil
}

// Set 把 time.Duration 转成 redis provider 所需的秒级 TTL。
func (c *redisCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.r.Set(ctx, key, value, int(ttl.Seconds()))
}

func (c *redisCache) Del(ctx context.Context, key string) error {
	return c.r.Del(ctx, key)
}

func (c *redisCache) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	return c.r.MGet(ctx, keys...)
}
