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
// - 需要项目引入 go.etcd.io/etcd/client/v3 依赖。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.etcd" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, _ := getConfigSourceConfig(c)
		return NewSource(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getConfigSourceConfig 从容器获取配置源配置。
func getConfigSourceConfig(c contract.Container) (*contract.ConfigSourceConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("config: invalid config service")
	}

	sourceCfg := &contract.ConfigSourceConfig{
		Type: contract.ConfigSourceEtcd,
	}

	endpoints := configprovider.GetStringSliceAny(cfg,
		"configsource.etcd.endpoints",
		"config_source.etcd.endpoints",
		"config_source.etcd_endpoints",
	)
	if len(endpoints) == 0 {
		endpoints = []string{"localhost:2379"} // 默认地址
	}
	sourceCfg.EtcdEndpoints = endpoints

	if etcdPath := configprovider.GetStringAny(cfg,
		"configsource.etcd.path",
		"config_source.etcd.path",
		"config_source.etcd_path",
	); etcdPath != "" {
		sourceCfg.EtcdPath = etcdPath
	} else {
		sourceCfg.EtcdPath = "/config/" // 默认路径
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

// Source 是 etcd 配置源实现。
//
// 中文说明：
// - 使用 etcd KV 存储配置；
// - 支持配置监听（通过 etcd Watch 实现）；
// - 支持动态写入。
type Source struct {
	cfg    *contract.ConfigSourceConfig
	client *clientv3.Client

	// 监听器管理
	watchers sync.Map // map[string]*etcdWatcher
	mu       sync.Mutex
	closed   bool
}

// NewSource 创建 etcd 配置源。
func NewSource(cfg *contract.ConfigSourceConfig) (*Source, error) {
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

	return &Source{
		cfg:    cfg,
		client: client,
	}, nil
}

// Load 从 etcd 加载配置。
//
// 中文说明：
// - 读取 EtcdPath 下的所有键值；
// - 支持嵌套路径（如 /config/app/database）；
// - 返回合并后的配置 map。
func (s *Source) Load(ctx context.Context) (map[string]any, error) {
	resp, err := s.client.Get(ctx, s.cfg.EtcdPath, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("configsource.etcd: get keys failed: %w", err)
	}

	result := make(map[string]any)
	for _, kv := range resp.Kvs {
		// 移除路径前缀
		key := strings.TrimPrefix(string(kv.Key), s.cfg.EtcdPath)
		key = strings.TrimPrefix(key, "/")

		// 解析值（支持 JSON）
		var value any
		if err := json.Unmarshal(kv.Value, &value); err == nil {
			// JSON 格式
			setNestedValue(result, key, value)
		} else {
			// 非 JSON 格式，作为字符串
			setNestedValue(result, key, string(kv.Value))
		}
	}

	return result, nil
}

// Get 获取单个配置项。
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

// Set 设置单个配置项。
//
// 中文说明：
// - 写入 etcd KV 存储；
// - 支持 JSON 格式存储复杂对象。
func (s *Source) Set(ctx context.Context, key string, value any) error {
	fullKey := path.Join(s.cfg.EtcdPath, key)

	// 序列化值
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

// Watch 监听配置变化。
//
// 中文说明：
// - 使用 etcd Watch API 实现监听；
// - 当配置发生变化时触发回调。
func (s *Source) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, errors.New("configsource.etcd: source closed")
	}

	// 检查是否已存在监听器
	if cached, ok := s.watchers.Load(key); ok {
		return cached.(*etcdWatcher), nil
	}

	fullKey := path.Join(s.cfg.EtcdPath, key)
	watcher := &etcdWatcher{
		source: s,
		key:    fullKey,
		ctx:    ctx,
		cancel: func() {}, // 初始化空函数
	}

	s.watchers.Store(key, watcher)

	// 启动监听
	watcher.startWatch(ctx)

	return watcher, nil
}

// Close 关闭 etcd 配置源。
func (s *Source) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}
	s.closed = true

	// 停止所有监听器
	s.watchers.Range(func(key, value any) bool {
		watcher := value.(*etcdWatcher)
		watcher.Stop()
		return true
	})

	return s.client.Close()
}

// etcdWatcher 是 etcd 配置监听器。
type etcdWatcher struct {
	source   *Source
	key      string
	ctx      context.Context
	cancel   context.CancelFunc
	watchCh  clientv3.WatchChan
	callbacks sync.Map // map[string]func(any)
}

// OnChange 注册配置变化回调。
func (w *etcdWatcher) OnChange(key string, callback func(value any)) {
	w.callbacks.Store(key, callback)
}

// Stop 停止监听。
func (w *etcdWatcher) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}
	return nil
}

// startWatch 启动监听。
func (w *etcdWatcher) startWatch(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.watchCh = w.source.client.Watch(ctx, w.key)

	go w.watchLoop()
}

// watchLoop 监听循环。
//
// 中文说明：
// - 监听 etcd Watch 通道；
// - 当收到 PUT 事件时触发回调。
func (w *etcdWatcher) watchLoop() {
	for resp := range w.watchCh {
		for _, ev := range resp.Events {
			if ev.Type == clientv3.EventTypePut {
				// 解析新值
				var value any
				if err := json.Unmarshal(ev.Kv.Value, &value); err == nil {
					// JSON 格式
				} else {
					value = string(ev.Kv.Value)
				}

				// 触发所有回调
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

// setNestedValue 设置嵌套 map 值。
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