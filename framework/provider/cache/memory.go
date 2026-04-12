package cache

import (
	"context"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
)

// memoryStore 是一个带 TTL 的轻量级内存缓存实现。
//
// 中文说明：
// - 适合本地开发、单元测试、单进程场景。
// - 数据只存在当前进程内，应用重启后会全部丢失。
// - 过期数据采用“访问时清理”策略，不额外启动后台清扫协程，保持实现简单。
type memoryStore struct {
	mu sync.RWMutex
	m  map[string]memItem
}

type memItem struct {
	v       string
	expired time.Time // zero means no expiration
}

func newMemoryStore() *memoryStore {
	return &memoryStore{m: make(map[string]memItem)}
}

// Get 读取缓存项，并在发现过期时顺手清理。
//
// 中文说明：
// - 先用读锁拿快照，降低高频读取时的锁竞争。
// - 如果 key 不存在，直接返回 ErrCacheMiss。
// - 如果 key 已过期，再升级为写锁删除该项，保证后续读取不会再命中过期数据。
func (s *memoryStore) Get(ctx context.Context, key string) (string, error) {
	_ = ctx

	s.mu.RLock()
	item, ok := s.m[key]
	s.mu.RUnlock()
	if !ok {
		return "", contract.ErrCacheMiss
	}
	if !item.expired.IsZero() && time.Now().After(item.expired) {
		s.mu.Lock()
		delete(s.m, key)
		s.mu.Unlock()
		return "", contract.ErrCacheMiss
	}
	return item.v, nil
}

func (s *memoryStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	// 中文说明：
	// - ttl<=0 时，exp 保持零值，表示该 key 永不过期。
	// - 写入时直接覆盖旧值，语义与大多数缓存系统保持一致。
	s.mu.Lock()
	s.m[key] = memItem{v: value, expired: exp}
	s.mu.Unlock()
	return nil
}

// Del 删除单个 key，即使 key 不存在也视为成功。
func (s *memoryStore) Del(ctx context.Context, key string) error {
	_ = ctx
	s.mu.Lock()
	delete(s.m, key)
	s.mu.Unlock()
	return nil
}

func (s *memoryStore) MGet(ctx context.Context, keys ...string) (map[string]string, error) {
	_ = ctx

	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v, err := s.Get(ctx, k)
		if err != nil {
			if err == contract.ErrCacheMiss {
				continue
			}
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}
