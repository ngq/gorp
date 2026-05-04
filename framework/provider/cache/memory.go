package cache

import (
	"context"
	"sync"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
)

type memoryStore struct {
	mu sync.RWMutex
	m  map[string]memItem
}

type memItem struct {
	v       string
	expired time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{m: make(map[string]memItem)}
}

func (s *memoryStore) Get(ctx context.Context, key string) (string, error) {
	_ = ctx

	s.mu.RLock()
	item, ok := s.m[key]
	s.mu.RUnlock()
	if !ok {
		return "", datacontract.ErrCacheMiss
	}
	if !item.expired.IsZero() && time.Now().After(item.expired) {
		s.mu.Lock()
		delete(s.m, key)
		s.mu.Unlock()
		return "", datacontract.ErrCacheMiss
	}
	return item.v, nil
}

func (s *memoryStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	_ = ctx

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	s.mu.Lock()
	s.m[key] = memItem{v: value, expired: exp}
	s.mu.Unlock()
	return nil
}

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
			if err == datacontract.ErrCacheMiss {
				continue
			}
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}
