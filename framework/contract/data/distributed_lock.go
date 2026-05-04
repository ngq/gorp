package data

import (
	"context"
	"time"
)

const DistributedLockKey = "framework.distributed_lock"

type DistributedLock interface {
	Lock(ctx context.Context, key string, ttl time.Duration) error
	TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	Renew(ctx context.Context, key string, ttl time.Duration) error
	IsLocked(ctx context.Context, key string) (bool, error)
	WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error
}

type DistributedLockConfig struct {
	Type string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	EtcdEndpoints []string
	EtcdUsername  string
	EtcdPassword  string

	DefaultTTL    time.Duration
	RetryInterval time.Duration
	MaxRetry      int
	KeyPrefix     string
}
