package noop

import (
	"context"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "dlock.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{datacontract.DistributedLockKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.DistributedLockKey, func(c runtimecontract.Container) (any, error) {
		return &noopLock{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

type noopLock struct {
	locks sync.Map
}

func (l *noopLock) Lock(ctx context.Context, key string, ttl time.Duration) error {
	_ = ttl
	lock := l.getOrCreateLock(key)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	lock.Lock()
	return nil
}

func (l *noopLock) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	_ = ttl
	lock := l.getOrCreateLock(key)

	locked := make(chan struct{})
	go func() {
		lock.Lock()
		close(locked)
	}()

	select {
	case <-locked:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (l *noopLock) Unlock(ctx context.Context, key string) error {
	_ = ctx
	if lock, ok := l.locks.Load(key); ok {
		lock.(*sync.Mutex).Unlock()
	}
	return nil
}

func (l *noopLock) Renew(ctx context.Context, key string, ttl time.Duration) error {
	_ = ctx
	_ = key
	_ = ttl
	return nil
}

func (l *noopLock) IsLocked(ctx context.Context, key string) (bool, error) {
	lock := l.getOrCreateLock(key)

	locked := make(chan bool)
	go func() {
		if lock.TryLock() {
			lock.Unlock()
			locked <- false
		} else {
			locked <- true
		}
	}()

	select {
	case result := <-locked:
		return result, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (l *noopLock) WithLock(ctx context.Context, key string, ttl time.Duration, fn func() error) error {
	if err := l.Lock(ctx, key, ttl); err != nil {
		return err
	}
	defer l.Unlock(ctx, key)
	return fn()
}

func (l *noopLock) getOrCreateLock(key string) *sync.Mutex {
	if lock, ok := l.locks.Load(key); ok {
		return lock.(*sync.Mutex)
	}

	lock := &sync.Mutex{}
	l.locks.Store(key, lock)
	return lock
}
