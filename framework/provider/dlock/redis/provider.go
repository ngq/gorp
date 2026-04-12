package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	"github.com/redis/go-redis/v9"
)

// Provider 提供 Redis 分布式锁实现。
//
// 中文说明：
// - 使用 Redis SET NX EX 原子命令实现；
// - 支持锁续约（看门狗）；
// - 只能释放自己持有的锁；
// - 需要项目引入 github.com/redis/go-redis/v9 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "dlock.redis" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.DistributedLockKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DistributedLockKey, func(c contract.Container) (any, error) {
		cfg, err := getLockConfig(c)
		if err != nil {
			return nil, err
		}
		return NewLock(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getLockConfig 从容器获取分布式锁配置。
func getLockConfig(c contract.Container) (*contract.DistributedLockConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("dlock: invalid config service")
	}

	lockCfg := &contract.DistributedLockConfig{
		Type:          "redis",
		DefaultTTL:    30 * time.Second,
		RetryInterval: 100 * time.Millisecond,
		MaxRetry:      50,
		KeyPrefix:     "lock:",
	}

	// Redis 配置
	if addr := cfg.GetString("dlock.redis_addr"); addr != "" {
		lockCfg.RedisAddr = addr
	} else {
		lockCfg.RedisAddr = "localhost:6379"
	}
	if password := cfg.GetString("dlock.redis_password"); password != "" {
		lockCfg.RedisPassword = password
	}
	if db := cfg.GetInt("dlock.redis_db"); db > 0 {
		lockCfg.RedisDB = db
	}

	// 通用配置
	if prefix := cfg.GetString("dlock.key_prefix"); prefix != "" {
		lockCfg.KeyPrefix = prefix
	}

	return lockCfg, nil
}

// Lock 是 Redis 分布式锁实现。
type Lock struct {
	cfg    *contract.DistributedLockConfig
	client *redis.Client

	// 持有的锁
	heldLocks sync.Map // map[string]*heldLock

	// 看门狗
	watchdog sync.Map // map[string]context.CancelFunc

	closed bool
	mu     sync.Mutex
}

// heldLock 持有的锁信息。
type heldLock struct {
	value string    // 锁的唯一标识
	token string    // 随机令牌
	held  time.Time // 获取时间
}

// NewLock 创建 Redis 分布式锁。
func NewLock(cfg *contract.DistributedLockConfig) (*Lock, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("dlock.redis: connect failed: %w", err)
	}

	return &Lock{
		cfg:    cfg,
		client: client,
	}, nil
}

// Lock 获取锁（阻塞）。
func (l *Lock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key

	// 生成唯一令牌
	token := generateToken()

	// 重试获取锁
	for i := 0; i < l.cfg.MaxRetry; i++ {
		// 尝试获取锁
		ok, err := l.tryLock(ctx, fullKey, token, ttl)
		if err != nil {
			return err
		}
		if ok {
			// 记录持有的锁
			l.heldLocks.Store(fullKey, &heldLock{
				value: fullKey,
				token: token,
				held:  time.Now(),
			})

			// 启动看门狗
			l.startWatchdog(ctx, fullKey, token, ttl)

			return nil
		}

		// 等待重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.cfg.RetryInterval):
		}
	}

	return errors.New("dlock: failed to acquire lock after max retries")
}

// TryLock 尝试获取锁（非阻塞）。
func (l *Lock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	fullKey := l.cfg.KeyPrefix + key
	token := generateToken()

	ok, err := l.tryLock(ctx, fullKey, token, ttl)
	if err != nil {
		return false, err
	}

	if ok {
		// 记录持有的锁
		l.heldLocks.Store(fullKey, &heldLock{
			value: fullKey,
			token: token,
			held:  time.Now(),
		})

		// 启动看门狗
		l.startWatchdog(ctx, fullKey, token, ttl)
	}

	return ok, nil
}

// tryLock 尝试获取锁。
func (l *Lock) tryLock(ctx context.Context, key string, token string, ttl time.Duration) (bool, error) {
	// 使用 SET NX EX 原子命令
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("dlock: setnx failed: %w", err)
	}
	return ok, nil
}

// Unlock 释放锁。
func (l *Lock) Unlock(ctx context.Context, key string) error {
	fullKey := l.cfg.KeyPrefix + key

	// 获取锁信息
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}

	// 停止看门狗
	l.stopWatchdog(fullKey)

	// 使用 Lua 脚本原子释放锁
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
	`

	result, err := l.client.Eval(ctx, script, []string{fullKey}, held.(*heldLock).token).Result()
	if err != nil {
		return fmt.Errorf("dlock: unlock failed: %w", err)
	}

	// 删除持有记录
	l.heldLocks.Delete(fullKey)

	if result.(int64) == 0 {
		return errors.New("dlock: lock not owned")
	}

	return nil
}

// Renew 续约锁。
func (l *Lock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key

	// 获取锁信息
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}

	// 使用 Lua 脚本原子续约
	script := `
	if redis.call("get", KEYS[1]) == ARGV[1] then
		return redis.call("expire", KEYS[1], ARGV[2])
	else
		return 0
	end
	`

	result, err := l.client.Eval(ctx, script, []string{fullKey}, held.(*heldLock).token, int(ttl/time.Second)).Result()
	if err != nil {
		return fmt.Errorf("dlock: renew failed: %w", err)
	}

	if result.(int64) == 0 {
		return errors.New("dlock: lock not owned")
	}

	return nil
}

// IsLocked 检查锁是否被持有。
func (l *Lock) IsLocked(ctx context.Context, key string) (bool, error) {
	fullKey := l.cfg.KeyPrefix + key

	result, err := l.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}

	return result > 0, nil
}

// WithLock 获取锁并执行函数。
func (l *Lock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if err := l.Lock(ctx, key, ttl); err != nil {
		return err
	}
	defer l.Unlock(ctx, key)

	return fn()
}

// startWatchdog 启动看门狗（自动续约）。
func (l *Lock) startWatchdog(ctx context.Context, key string, token string, ttl time.Duration) {
	// 续约间隔为 TTL 的 1/3
	interval := ttl / 3

	ctx, cancel := context.WithCancel(context.Background())
	l.watchdog.Store(key, cancel)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 自动续约
				renewCtx, cancel := context.WithTimeout(context.Background(), time.Second)
				_ = l.Renew(renewCtx, key, ttl)
				cancel()
			}
		}
	}()
}

// stopWatchdog 停止看门狗。
func (l *Lock) stopWatchdog(key string) {
	if cancel, ok := l.watchdog.Load(key); ok {
		cancel.(context.CancelFunc)()
		l.watchdog.Delete(key)
	}
}

// generateToken 生成随机令牌。
func generateToken() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}