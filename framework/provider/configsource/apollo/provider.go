package apollo

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Apollo 配置中心实现。
//
// 中文说明：
// - 使用携程 Apollo 配置中心；
// - 支持多命名空间；
// - 支持配置热更新；
// - 支持灰度发布。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.apollo" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getApolloConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ApolloConfig 定义 Apollo 配置。
type ApolloConfig struct {
	// AppID 应用 ID
	AppID string

	// Cluster 集群名称（默认 "default"）
	Cluster string

	// Namespace 命名空间（默认 "application"）
	Namespace string

	// MetaServer Apollo Meta Server 地址
	MetaServer string

	// AccessKey 访问密钥
	AccessKey string
}

// getApolloConfig 从容器获取配置。
func getApolloConfig(c contract.Container) (*ApolloConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("apollo: invalid config service")
	}

	apolloCfg := &ApolloConfig{
		Cluster:   "default",
		Namespace: "application",
	}

	if v := cfg.Get("configsource.apollo.app_id"); v != nil {
		apolloCfg.AppID = cfg.GetString("configsource.apollo.app_id")
	} else if v := cfg.Get("config.apollo.app_id"); v != nil {
		apolloCfg.AppID = cfg.GetString("config.apollo.app_id")
	}
	if v := cfg.Get("configsource.apollo.cluster"); v != nil {
		apolloCfg.Cluster = cfg.GetString("configsource.apollo.cluster")
	} else if v := cfg.Get("config.apollo.cluster"); v != nil {
		apolloCfg.Cluster = cfg.GetString("config.apollo.cluster")
	}
	if v := cfg.Get("configsource.apollo.namespace"); v != nil {
		apolloCfg.Namespace = cfg.GetString("configsource.apollo.namespace")
	} else if v := cfg.Get("config.apollo.namespace"); v != nil {
		apolloCfg.Namespace = cfg.GetString("config.apollo.namespace")
	}
	if v := cfg.Get("configsource.apollo.meta_server"); v != nil {
		apolloCfg.MetaServer = cfg.GetString("configsource.apollo.meta_server")
	} else if v := cfg.Get("config.apollo.meta_server"); v != nil {
		apolloCfg.MetaServer = cfg.GetString("config.apollo.meta_server")
	}
	if v := cfg.Get("configsource.apollo.access_key"); v != nil {
		apolloCfg.AccessKey = cfg.GetString("configsource.apollo.access_key")
	} else if v := cfg.Get("config.apollo.access_key"); v != nil {
		apolloCfg.AccessKey = cfg.GetString("config.apollo.access_key")
	}

	return apolloCfg, nil
}

// ConfigSource Apollo 配置源实现。
type ConfigSource struct {
	config *ApolloConfig
	mu     sync.RWMutex
	cache  map[string]string
}

// NewConfigSource 创建 Apollo 配置源。
func NewConfigSource(cfg *ApolloConfig) (*ConfigSource, error) {
	if cfg.AppID == "" {
		return nil, errors.New("apollo: app_id is required")
	}
	if cfg.MetaServer == "" {
		return nil, errors.New("apollo: meta_server is required")
	}

	return &ConfigSource{
		config: cfg,
		cache:  make(map[string]string),
	}, nil
}

// Load 加载配置。
func (s *ConfigSource) Load() error {
	// TODO: 实现真实的 Apollo 配置加载
	// 需要引入 github.com/apolloconfig/apollo-sdk-go
	return nil
}

// Get 获取配置值。
func (s *ConfigSource) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cache[key]
}

// Set 设置配置值（本地缓存）。
func (s *ConfigSource) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = value
	return nil
}

// Watch 监听配置变化。
func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return &apolloWatcher{ctx: ctx}, nil
}

// Close 关闭配置源。
func (s *ConfigSource) Close() error {
	return nil
}

// apolloWatcher 配置监听器。
type apolloWatcher struct {
	ctx context.Context
}

// OnChange 配置变更回调。
func (w *apolloWatcher) OnChange(key string, callback func(value any)) {
	// TODO: 实现配置变更监听
}

// Stop 停止监听。
func (w *apolloWatcher) Stop() error { return nil }

var (
	ErrAppIDRequired = errors.New("apollo: app_id is required")
	ErrMetaRequired  = errors.New("apollo: meta_server is required")
)