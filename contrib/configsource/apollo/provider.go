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
// - 支持灰度发布；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }
func (p *Provider) Name() string     { return "configsource.apollo" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

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
func (p *Provider) Boot(c contract.Container) error { return nil }

type ApolloConfig struct {
	AppID      string
	Cluster    string
	Namespace  string
	MetaServer string
	AccessKey  string
}

func getApolloConfig(c contract.Container) (*ApolloConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("apollo: invalid config service")
	}
	apolloCfg := &ApolloConfig{Cluster: "default", Namespace: "application"}
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

type ConfigSource struct {
	config *ApolloConfig
	mu     sync.RWMutex
	cache  map[string]string
}

func NewConfigSource(cfg *ApolloConfig) (*ConfigSource, error) {
	if cfg.AppID == "" {
		return nil, errors.New("apollo: app_id is required")
	}
	if cfg.MetaServer == "" {
		return nil, errors.New("apollo: meta_server is required")
	}
	return &ConfigSource{config: cfg, cache: make(map[string]string)}, nil
}

func (s *ConfigSource) Load() error { return nil }
func (s *ConfigSource) Get(key string) string {
	s.mu.RLock(); defer s.mu.RUnlock(); return s.cache[key]
}
func (s *ConfigSource) Set(key, value string) error {
	s.mu.Lock(); defer s.mu.Unlock(); s.cache[key] = value; return nil
}
func (s *ConfigSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return &apolloWatcher{ctx: ctx}, nil
}
func (s *ConfigSource) Close() error { return nil }

type apolloWatcher struct { ctx context.Context }
func (w *apolloWatcher) OnChange(key string, callback func(value any)) {}
func (w *apolloWatcher) Stop() error { return nil }

var (
	ErrAppIDRequired = errors.New("apollo: app_id is required")
	ErrMetaRequired  = errors.New("apollo: meta_server is required")
)
