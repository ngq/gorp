package contract

import (
	"context"
	"time"
)

const (
	// DistributedLockKey 是分布式锁在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于分布式环境下的并发控制；
	// - 支持 Redis、etcd 等多种实现；
	// - noop 实现空操作，单体项目零依赖。
	DistributedLockKey = "framework.distributed_lock"
)

// DistributedLock 分布式锁接口。
//
// 中文说明：
// - 统一的分布式锁抽象；
// - 支持获取、释放、续约；
// - 支持可重入锁。
type DistributedLock interface {
	// Lock 获取锁（阻塞）。
	//
	// 中文说明：
	// - key: 锁的键名；
	// - ttl: 锁的存活时间；
	// - 阻塞直到获取锁或上下文取消。
	Lock(ctx context.Context, key string, ttl time.Duration) error

	// TryLock 尝试获取锁（非阻塞）。
	//
	// 中文说明：
	// - 如果锁可用则立即获取；
	// - 返回是否成功获取。
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Unlock 释放锁。
	//
	// 中文说明：
	// - 释放当前持有的锁；
	// - 只能释放自己持有的锁。
	Unlock(ctx context.Context, key string) error

	// Renew 续约锁。
	//
	// 中文说明：
	// - 延长锁的存活时间；
	// - 只能续约自己持有的锁。
	Renew(ctx context.Context, key string, ttl time.Duration) error

	// IsLocked 检查锁是否被持有。
	IsLocked(ctx context.Context, key string) (bool, error)

	// WithLock 获取锁并执行函数。
	//
	// 中文说明：
	// - 自动获取和释放锁；
	// - 执行完成后自动释放。
	WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error
}

// DistributedLockConfig 分布式锁配置。
type DistributedLockConfig struct {
	// Type 锁类型：noop/redis/etcd
	Type string

	// Redis 配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// etcd 配置
	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string

	// 通用配置
	DefaultTTL    time.Duration // 默认锁存活时间
	RetryInterval time.Duration // 获取锁重试间隔
	MaxRetry      int           // 最大重试次数
	KeyPrefix     string        // 键前缀
}