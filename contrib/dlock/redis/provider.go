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
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "dlock.redis" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.DistributedLockKey} }

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

func (p *Provider) Boot(c contract.Container) error { return nil }

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
	if addr := cfg.GetString("distributed_lock.redis.addr"); addr != "" {
		lockCfg.RedisAddr = addr
	} else {
		lockCfg.RedisAddr = "localhost:6379"
	}
	if password := cfg.GetString("distributed_lock.redis.password"); password != "" {
		lockCfg.RedisPassword = password
	}
	if db := cfg.GetInt("distributed_lock.redis.db"); db > 0 {
		lockCfg.RedisDB = db
	}
	if prefix := cfg.GetString("distributed_lock.key_prefix"); prefix != "" {
		lockCfg.KeyPrefix = prefix
	} else if prefix := cfg.GetString("dlock.key_prefix"); prefix != "" {
		lockCfg.KeyPrefix = prefix
	}
	return lockCfg, nil
}

// Lock 是 Redis 分布式锁实现。
type Lock struct {
	cfg      *contract.DistributedLockConfig
	client   *redis.Client
	heldLocks sync.Map
	watchdog sync.Map
	closed   bool
	mu       sync.Mutex
}

type heldLock struct {
	value string
	token string
	held  time.Time
}

// NewLock 创建 Redis 分布式锁。
func NewLock(cfg *contract.DistributedLockConfig) (*Lock, error) {
	client := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("dlock.redis: connect failed: %w", err)
	}
	return &Lock{cfg: cfg, client: client}, nil
}

// Lock 获取锁（阻塞）。
func (l *Lock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key
	token := generateToken()
	for i := 0; i < l.cfg.MaxRetry; i++ {
		ok, err := l.tryLock(ctx, fullKey, token, ttl)
		if err != nil {
			return err
		}
		if ok {
			l.heldLocks.Store(fullKey, &heldLock{value: fullKey, token: token, held: time.Now()})
			l.startWatchdog(ctx, fullKey, token, ttl)
			return nil
		}
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
		l.heldLocks.Store(fullKey, &heldLock{value: fullKey, token: token, held: time.Now()})
		l.startWatchdog(ctx, fullKey, token, ttl)
	}
	return ok, nil
}

func (l *Lock) tryLock(ctx context.Context, key string, token string, ttl time.Duration) (bool, error) {
	ok, err := l.client.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("dlock: setnx failed: %w", err)
	}
	return ok, nil
}

// Unlock 释放锁。
func (l *Lock) Unlock(ctx context.Context, key string) error {
	fullKey := l.cfg.KeyPrefix + key
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}
	l.stopWatchdog(fullKey)
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
	l.heldLocks.Delete(fullKey)
	if result.(int64) == 0 {
		return errors.New("dlock: lock not owned")
	}
	return nil
}

// Renew 续约锁。
func (l *Lock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := l.cfg.KeyPrefix + key
	held, ok := l.heldLocks.Load(fullKey)
	if !ok {
		return errors.New("dlock: lock not held")
	}
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
	if l == nil || l.client == nil {
		return false, errors.New("dlock: client not initialized")
	}
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

func (l *Lock) startWatchdog(ctx context.Context, key string, token string, ttl time.Duration) {
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
				renewCtx, cancel := context.WithTimeout(context.Background(), time.Second)
				_ = l.Renew(renewCtx, key, ttl)
				cancel()
			}
		}
	}()
}

func (l *Lock) stopWatchdog(key string) {
	if cancel, ok := l.watchdog.Load(key); ok {
		cancel.(context.CancelFunc)()
		l.watchdog.Delete(key)
	}
}

func generateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
