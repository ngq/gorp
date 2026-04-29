package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/ngq/gorp/framework/contract"
	configprovider "github.com/ngq/gorp/framework/provider/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Provider 提供 etcd 配置源实现。
//
// 中文说明：
// - 从 etcd KV 存储读取配置；
// - 支持配置热更新（通过 Watch 监听变化）；
// - 支持动态写入配置；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "configsource.etcd" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfigSourceConfig(c)
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
	sourceCfg := &contract.ConfigSourceConfig{Type: contract.ConfigSourceEtcd}
	endpoints := configprovider.GetStringSliceAny(cfg,
		"configsource.etcd.endpoints",
		"config_source.etcd.endpoints",
		"config_source.etcd_endpoints",
	)
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	sourceCfg.EtcdEndpoints = endpoints
	if etcdPath := configprovider.GetStringAny(cfg,
		"configsource.etcd.path",
		"config_source.etcd.path",
		"config_source.etcd_path",
	); etcdPath != "" {
		sourceCfg.EtcdPath = etcdPath
	} else {
		sourceCfg.EtcdPath = "/config/"
	}
	if username := configprovider.GetStringAny(cfg,
		"configsource.etcd.username",
		"config_source.etcd.username",
		"config_source.etcd_username",
	); username != "" {
		sourceCfg.EtcdUsername = username
	}
	if password := configprovider.GetStringAny(cfg,
		"configsource.etcd.password",
		"config_source.etcd.password",
		"config_source.etcd_password",
	); password != "" {
		sourceCfg.EtcdPassword = password
	}
	return sourceCfg, nil
}

type Source struct {
	cfg      *contract.ConfigSourceConfig
	client   *clientv3.Client
	watchers sync.Map
	mu       sync.Mutex
	closed   bool
}

func NewSource(cfg *contract.ConfigSourceConfig) (*Source, error) {
	clientCfg := clientv3.Config{Endpoints: cfg.EtcdEndpoints, DialTimeout: 5 * time.Second}
	if cfg.EtcdUsername != "" && cfg.EtcdPassword != "" {
		clientCfg.Username = cfg.EtcdUsername
		clientCfg.Password = cfg.EtcdPassword
	}
	client, err := clientv3.New(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("configsource.etcd: create client failed: %w", err)
	}
	return &Source{cfg: cfg, client: client}, nil
}

func (s *Source) Load(ctx context.Context) (map[string]any, error) {
	resp, err := s.client.Get(ctx, s.cfg.EtcdPath, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("configsource.etcd: get keys failed: %w", err)
	}
	result := make(map[string]any)
	for _, kv := range resp.Kvs {
		key := strings.TrimPrefix(string(kv.Key), s.cfg.EtcdPath)
		key = strings.TrimPrefix(key, "/")
		var value any
		if err := json.Unmarshal(kv.Value, &value); err == nil {
			setNestedValue(result, key, value)
		} else {
			setNestedValue(result, key, string(kv.Value))
		}
	}
	return result, nil
}

func (s *Source) Get(ctx context.Context, key string) (any, error) {
	fullKey := path.Join(s.cfg.EtcdPath, key)
	resp, err := s.client.Get(ctx, fullKey)
	if err != nil {
		return nil, fmt.Errorf("configsource.etcd: get key failed: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("configsource.etcd: key %s not found", key)
	}
	var value any
	if err := json.Unmarshal(resp.Kvs[0].Value, &value); err == nil {
		return value, nil
	}
	return string(resp.Kvs[0].Value), nil
}

func (s *Source) Set(ctx context.Context, key string, value any) error {
	fullKey := path.Join(s.cfg.EtcdPath, key)
	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		data, _ = json.Marshal(v)
	}
	_, err := s.client.Put(ctx, fullKey, string(data))
	if err != nil {
		return fmt.Errorf("configsource.etcd: put key failed: %w", err)
	}
	return nil
}

func (s *Source) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("configsource.etcd: source closed")
	}
	if cached, ok := s.watchers.Load(key); ok {
		return cached.(*etcdWatcher), nil
	}
	fullKey := path.Join(s.cfg.EtcdPath, key)
	watcher := &etcdWatcher{source: s, key: fullKey, ctx: ctx, cancel: func() {}}
	s.watchers.Store(key, watcher)
	watcher.startWatch(ctx)
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
		watcher := value.(*etcdWatcher)
		watcher.Stop()
		return true
	})
	return s.client.Close()
}

type etcdWatcher struct {
	source    *Source
	key       string
	ctx       context.Context
	cancel    context.CancelFunc
	callbacks sync.Map
}

func (w *etcdWatcher) OnChange(key string, callback func(value any)) { w.callbacks.Store(key, callback) }
func (w *etcdWatcher) Stop() error { w.cancel(); return nil }

func (w *etcdWatcher) startWatch(ctx context.Context) {
	watchCtx, cancel := context.WithCancel(ctx)
	w.cancel = cancel
	go func() {
		watchCh := w.source.client.Watch(watchCtx, w.key)
		for resp := range watchCh {
			for _, event := range resp.Events {
				var value any
				if event.Kv != nil {
					if err := json.Unmarshal(event.Kv.Value, &value); err != nil {
						value = string(event.Kv.Value)
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
	}()
}

func setNestedValue(result map[string]any, key string, value any) {
	keys := strings.Split(key, "/")
	current := result
	for i, k := range keys {
		if k == "" {
			continue
		}
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
