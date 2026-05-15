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

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/ngq/gorp/contrib/internal/baseconfigsource"
	internalnative "github.com/ngq/gorp/contrib/internal/native"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// Provider implements runtimecontract.ServiceProvider for etcd config source.
type Provider struct {
	baseconfigsource.BaseConfigSourceProvider
}

// NewProvider creates a new etcd config source provider.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "configsource.etcd"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getConfigSourceConfig(c)
	}
	p.NewSource = func(cfg any) (datacontract.ConfigSource, error) {
		return NewSource(cfg.(*datacontract.ConfigSourceConfig))
	}
	return p
}

func getConfigSourceConfig(c runtimecontract.Container) (*datacontract.ConfigSourceConfig, error) {
	cfg, err := baseconfigsource.ReadConfig(c)
	if err != nil {
		return nil, err
	}

	sourceCfg := &datacontract.ConfigSourceConfig{Type: datacontract.ConfigSourceEtcd}
	endpoints := configprovider.GetStringSliceAny(cfg, "configsource.etcd.endpoints", "config_source.etcd.endpoints", "config_source.etcd_endpoints")
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"}
	}
	sourceCfg.EtcdEndpoints = endpoints

	if etcdPath := configprovider.GetStringAny(cfg, "configsource.etcd.path", "config_source.etcd.path", "config_source.etcd_path"); etcdPath != "" {
		sourceCfg.EtcdPath = etcdPath
	} else {
		sourceCfg.EtcdPath = "/config/"
	}

	if username := configprovider.GetStringAny(cfg, "configsource.etcd.username", "config_source.etcd.username", "config_source.etcd_username"); username != "" {
		sourceCfg.EtcdUsername = username
	}
	if password := configprovider.GetStringAny(cfg, "configsource.etcd.password", "config_source.etcd.password", "config_source.etcd_password"); password != "" {
		sourceCfg.EtcdPassword = password
	}

	return sourceCfg, nil
}

type Source struct {
	cfg    *datacontract.ConfigSourceConfig
	client *clientv3.Client

	watchers sync.Map
	mu       sync.Mutex
	closed   bool
}

func NewSource(cfg *datacontract.ConfigSourceConfig) (*Source, error) {
	clientCfg := clientv3.Config{
		Endpoints:   cfg.EtcdEndpoints,
		DialTimeout: 5 * time.Second,
	}

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

func (s *Source) Watch(ctx context.Context, key string) (datacontract.ConfigWatcher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("configsource.etcd: source closed")
	}
	if cached, ok := s.watchers.Load(key); ok {
		return cached.(*etcdWatcher), nil
	}

	fullKey := path.Join(s.cfg.EtcdPath, key)
	watcher := &etcdWatcher{
		source: s,
		key:    fullKey,
		ctx:    ctx,
		cancel: func() {},
	}

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

func (s *Source) Underlying() any {
	return s.client
}

func (s *Source) As(target any) bool {
	return internalnative.As(s.client, target)
}

type etcdWatcher struct {
	source    *Source
	key       string
	ctx       context.Context
	cancelMu  sync.Mutex
	cancel    context.CancelFunc
	watchCh   clientv3.WatchChan
	callbacks sync.Map
}

func (w *etcdWatcher) OnChange(key string, callback func(value any)) {
	w.callbacks.Store(key, callback)
}

func (w *etcdWatcher) Stop() error {
	w.cancelMu.Lock()
	defer w.cancelMu.Unlock()
	if w.cancel != nil {
		w.cancel()
	}
	return nil
}

func (w *etcdWatcher) startWatch(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	w.cancelMu.Lock()
	w.cancel = cancel
	w.watchCh = w.source.client.Watch(ctx, w.key)
	w.cancelMu.Unlock()

	go w.watchLoop()
}

func (w *etcdWatcher) watchLoop() {
	for resp := range w.watchCh {
		for _, ev := range resp.Events {
			if ev.Type == clientv3.EventTypePut {
				var value any
				if err := json.Unmarshal(ev.Kv.Value, &value); err != nil {
					value = string(ev.Kv.Value)
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
}

func setNestedValue(result map[string]any, key string, value any) {
	keys := strings.Split(key, "/")
	if len(keys) == 0 {
		return
	}

	current := result
	for i, k := range keys {
		if k == "" {
			continue
		}
		if i == len(keys)-1 || i == len(keys)-2 && keys[len(keys)-1] == "" {
			current[k] = value
		} else {
			if _, exists := current[k]; !exists {
				current[k] = make(map[string]any)
			}
			if next, ok := current[k].(map[string]any); ok {
				current = next
			}
		}
	}
}
