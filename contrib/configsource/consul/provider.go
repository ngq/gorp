package consul

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	"github.com/hashicorp/consul/api"
)

// Provider 提供 Consul KV 配置源实现。
//
// 中文说明：
// - 从 Consul KV 存储读取配置；
// - 支持配置热更新（通过 Watch 监听变化）；
// - 支持动态写入配置；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.consul" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getConfigSourceConfig(c)
		if err != nil {
			return nil, err
		}
		return NewSource(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

func getConfigSourceConfig(c contract.Container) (*contract.ConfigSourceConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("config: invalid config service")
	}

	sourceCfg := &contract.ConfigSourceConfig{Type: contract.ConfigSourceConsul}

	if addr := configprovider.GetStringAny(cfg,
		"configsource.consul.addr",
		"config_source.consul.addr",
		"config_source.consul_addr",
	); addr != "" {
		sourceCfg.ConsulAddr = addr
	} else {
		sourceCfg.ConsulAddr = "localhost:8500"
	}

	if path := configprovider.GetStringAny(cfg,
		"configsource.consul.path",
		"config_source.consul.path",
		"config_source.consul_path",
	); path != "" {
		sourceCfg.ConsulPath = path
	} else {
		sourceCfg.ConsulPath = "config/"
	}

	if token := configprovider.GetStringAny(cfg,
		"configsource.consul.token",
		"config_source.consul.token",
		"config_source.consul_token",
	); token != "" {
		sourceCfg.ConsulToken = token
	}

	return sourceCfg, nil
}

type Source struct {
	cfg     *contract.ConfigSourceConfig
	client  *api.Client
	kv      *api.KV
	watchers sync.Map
	mu      sync.Mutex
	closed  bool
}

func NewSource(cfg *contract.ConfigSourceConfig) (*Source, error) {
	consulCfg := api.DefaultConfig()
	if cfg.ConsulAddr != "" {
		consulCfg.Address = cfg.ConsulAddr
	}
	if cfg.ConsulToken != "" {
		consulCfg.Token = cfg.ConsulToken
	}

	client, err := api.NewClient(consulCfg)
	if err != nil {
		return nil, fmt.Errorf("configsource.consul: create client failed: %w", err)
	}

	return &Source{cfg: cfg, client: client, kv: client.KV()}, nil
}

func (s *Source) Load(ctx context.Context) (map[string]any, error) {
	pairs, _, err := s.kv.List(s.cfg.ConsulPath, nil)
	if err != nil {
		return nil, fmt.Errorf("configsource.consul: list keys failed: %w", err)
	}

	result := make(map[string]any)
	for _, pair := range pairs {
		key := stringsTrimPrefix(pair.Key, s.cfg.ConsulPath)
		var value any
		if err := json.Unmarshal(pair.Value, &value); err == nil {
			setNestedValue(result, key, value)
		} else {
			setNestedValue(result, key, string(pair.Value))
		}
	}
	return result, nil
}

func (s *Source) Get(ctx context.Context, key string) (any, error) {
	fullKey := s.cfg.ConsulPath + key
	pair, _, err := s.kv.Get(fullKey, nil)
	if err != nil {
		return nil, fmt.Errorf("configsource.consul: get key failed: %w", err)
	}
	if pair == nil {
		return nil, fmt.Errorf("configsource.consul: key %s not found", key)
	}
	var value any
	if err := json.Unmarshal(pair.Value, &value); err == nil {
		return value, nil
	}
	return string(pair.Value), nil
}

func (s *Source) Set(ctx context.Context, key string, value any) error {
	fullKey := s.cfg.ConsulPath + key
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, _ = json.Marshal(v)
	}
	pair := &api.KVPair{Key: fullKey, Value: data}
	_, err := s.kv.Put(pair, nil)
	if err != nil {
		return fmt.Errorf("configsource.consul: put key failed: %w", err)
	}
	return nil
}

func (s *Source) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("configsource.consul: source closed")
	}
	if cached, ok := s.watchers.Load(key); ok {
		return cached.(*consulWatcher), nil
	}
	fullKey := s.cfg.ConsulPath + key
	watcher := &consulWatcher{source: s, key: fullKey, stopCh: make(chan struct{})}
	s.watchers.Store(key, watcher)
	go watcher.watchLoop(ctx)
	return watcher, nil
}

func (s *Source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	s.watchers.Range(func(key, value any) bool {
		watcher := value.(*consulWatcher)
		watcher.Stop()
		return true
	})
	return nil
}

type consulWatcher struct {
	source    *Source
	key       string
	stopCh    chan struct{}
	callbacks sync.Map
}

func (w *consulWatcher) OnChange(key string, callback func(value any)) { w.callbacks.Store(key, callback) }
func (w *consulWatcher) Stop() error {
	select {
	case <-w.stopCh:
	default:
		close(w.stopCh)
	}
	return nil
}

func (w *consulWatcher) watchLoop(ctx context.Context) {
	var lastIndex uint64
	for {
		select {
		case <-w.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}
		pair, meta, err := w.source.kv.Get(w.key, &api.QueryOptions{WaitIndex: lastIndex, WaitTime: 30 * time.Second})
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		if meta.LastIndex > lastIndex {
			lastIndex = meta.LastIndex
			var value any
			if pair != nil {
				if err := json.Unmarshal(pair.Value, &value); err != nil {
					value = string(pair.Value)
				}
			}
			w.callbacks.Range(func(key, cb any) bool {
				if callback, ok := cb.(func(any)); ok {
					callback(value)
				}
				return true
			})
		}
	}
}

func setNestedValue(result map[string]any, key string, value any) {
	keys := stringsSplit(key, "/")
	if len(keys) == 0 {
		return
	}
	current := result
	for i, k := range keys {
		if i == len(keys)-1 {
			current[k] = value
		} else {
			if _, exists := current[k]; !exists {
				current[k] = make(map[string]any)
			}
			current = current[k].(map[string]any)
		}
	}
}

func stringsTrimPrefix(s, prefix string) string {
	for len(s) >= len(prefix) && s[:len(prefix)] == prefix {
		s = s[len(prefix):]
	}
	return s
}

func stringsSplit(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := []string{}
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}
