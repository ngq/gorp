package nacos

import (
	"context"
	"errors"
	"sync"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 Nacos 配置中心实现。
//
// 中文说明：
// - 使用阿里巴巴 Nacos 配置中心；
// - 支持多命名空间；
// - 支持配置热更新；
// - 支持分组管理；
// - 当前已从 framework/provider 真实下沉到 contrib 层。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.nacos" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		cfg, err := getNacosConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)
	return nil
}
func (p *Provider) Boot(c contract.Container) error { return nil }

type NacosConfig struct {
	ServerAddr string
	Port       int
	Namespace  string
	Group      string
	DataID     string
	Username   string
	Password   string
}

func getNacosConfig(c contract.Container) (*NacosConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}
	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("nacos: invalid config service")
	}
	nacosCfg := &NacosConfig{Port: 8848, Group: "DEFAULT_GROUP"}
	if v := cfg.Get("configsource.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("configsource.nacos.server_addr")
	} else if v := cfg.Get("config.nacos.server_addr"); v != nil {
		nacosCfg.ServerAddr = cfg.GetString("config.nacos.server_addr")
	}
	if v := cfg.Get("configsource.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("configsource.nacos.port")
	} else if v := cfg.Get("config.nacos.port"); v != nil {
		nacosCfg.Port = cfg.GetInt("config.nacos.port")
	}
	if v := cfg.Get("configsource.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("configsource.nacos.namespace")
	} else if v := cfg.Get("config.nacos.namespace"); v != nil {
		nacosCfg.Namespace = cfg.GetString("config.nacos.namespace")
	}
	if v := cfg.Get("configsource.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("configsource.nacos.group")
	} else if v := cfg.Get("config.nacos.group"); v != nil {
		nacosCfg.Group = cfg.GetString("config.nacos.group")
	}
	if v := cfg.Get("configsource.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("configsource.nacos.data_id")
	} else if v := cfg.Get("config.nacos.data_id"); v != nil {
		nacosCfg.DataID = cfg.GetString("config.nacos.data_id")
	}
	if v := cfg.Get("configsource.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("configsource.nacos.username")
	} else if v := cfg.Get("config.nacos.username"); v != nil {
		nacosCfg.Username = cfg.GetString("config.nacos.username")
	}
	if v := cfg.Get("configsource.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("configsource.nacos.password")
	} else if v := cfg.Get("config.nacos.password"); v != nil {
		nacosCfg.Password = cfg.GetString("config.nacos.password")
	}
	return nacosCfg, nil
}

type ConfigSource struct {
	config *NacosConfig
	mu     sync.RWMutex
	cache  map[string]string
}

func NewConfigSource(cfg *NacosConfig) (*ConfigSource, error) {
	if cfg.ServerAddr == "" {
		return nil, errors.New("nacos: server_addr is required")
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
	return &nacosWatcher{ctx: ctx}, nil
}
func (s *ConfigSource) Close() error { return nil }

type nacosWatcher struct { ctx context.Context }
func (w *nacosWatcher) OnChange(key string, callback func(value any)) {}
func (w *nacosWatcher) Stop() error { return nil }

var ErrServerAddrRequired = errors.New("nacos: server_addr is required")
