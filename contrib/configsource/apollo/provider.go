// Package apollo provides Apollo configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Apollo SDK integration.
//
// 本包提供 gorp 框架 Apollo 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Apollo SDK。
package apollo

import (
	"errors"
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Apollo 配置中心实现。
//
// 中文说明：
//   - 使用携程 Apollo 配置中心；
//   - 支持多命名空间；
//   - 支持配置热更新；
//   - 支持灰度发布。
//   - 当前状态：部分可用
//   - 说明：已完成 P2 第一版最小 HTTP 配置闭环，具备 Load / Watch 与 fake client 行为测试；
//     但当前仍是轮询桥接态，尚未进入完整 Apollo SDK 产品化能力。
type Provider struct{}

// NewProvider creates a new Apollo provider instance.
//
// NewProvider 创建新的 Apollo provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider identifier "configsource.apollo".
//
// Name 返回 provider 标识符 "configsource.apollo"。
func (p *Provider) Name() string  { return "configsource.apollo" }

// IsDefer returns true for lazy initialization.
//
// IsDefer 返回 true，延迟初始化。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the contract keys this provider satisfies.
//
// Provides 返回此 provider 满足的契约键。
func (p *Provider) Provides() []string {
	return []string{datacontract.ConfigSourceKey}
}

// Register binds the Apollo config source to the container.
//
// Register 将 Apollo 配置源绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := getApolloConfig(c)
		if err != nil {
			return nil, err
		}
		return NewConfigSource(cfg)
	}, true)

	return nil
}

// Boot does nothing for lazy providers.
//
// Boot 延迟 provider 不需要 boot 操作。
func (p *Provider) Boot(c runtimecontract.Container) error { return nil }

// ApolloConfig 定义 Apollo 配置。
type ApolloConfig struct {
	AppID              string
	Cluster            string
	Namespace          string
	MetaServer         string
	AccessKey          string
	PollInterval       time.Duration
	WatchRetryInterval time.Duration
}

// getApolloConfig extracts Apollo configuration from the container's config binding.
//
// getApolloConfig 从容器的 config binding 中提取 Apollo 配置。
func getApolloConfig(c runtimecontract.Container) (*ApolloConfig, error) {
	cfgAny, err := c.Make(datacontract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(datacontract.Config)
	if !ok {
		return nil, errors.New("apollo: invalid config service")
	}

	apolloCfg := &ApolloConfig{
		Cluster:            "default",
		Namespace:          "application",
		PollInterval:       defaultApolloPollInterval,
		WatchRetryInterval: time.Second,
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
	if v := cfg.Get("configsource.apollo.poll_interval_seconds"); v != nil {
		if seconds := cfg.GetInt("configsource.apollo.poll_interval_seconds"); seconds > 0 {
			apolloCfg.PollInterval = time.Duration(seconds) * time.Second
		}
	}
	if v := cfg.Get("configsource.apollo.watch_retry_interval_ms"); v != nil {
		if ms := cfg.GetInt("configsource.apollo.watch_retry_interval_ms"); ms > 0 {
			apolloCfg.WatchRetryInterval = time.Duration(ms) * time.Millisecond
		}
	}

	return apolloCfg, nil
}