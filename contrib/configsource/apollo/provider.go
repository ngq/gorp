// Package apollo provides Apollo configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Apollo SDK integration.
//
// 本包提供 gorp 框架 Apollo 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Apollo SDK。
package apollo

import (
	"time"

	"github.com/ngq/gorp/contrib/internal/baseconfigsource"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Apollo 配置中心实现。
type Provider struct {
	baseconfigsource.BaseConfigSourceProvider
}

// NewProvider creates a new Apollo provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "configsource.apollo"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getApolloConfig(c)
	}
	p.NewSource = func(cfg any) (datacontract.ConfigSource, error) {
		return NewConfigSource(cfg.(*ApolloConfig))
	}
	return p
}

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
// Uses GetStringFallback for single-path reading, eliminating the cfg.Get()+cfg.GetString() double-read.
//
// getApolloConfig 从容器的 config binding 中提取 Apollo 配置。
// 使用 GetStringFallback 单路径读取，消除 cfg.Get()+cfg.GetString() 双读冗余。
func getApolloConfig(c runtimecontract.Container) (*ApolloConfig, error) {
	cfg, err := baseconfigsource.ReadConfig(c)
	if err != nil {
		return nil, err
	}

	apolloCfg := &ApolloConfig{
		Cluster:            "default",
		Namespace:          "application",
		PollInterval:       defaultApolloPollInterval,
		WatchRetryInterval: time.Second,
	}

	apolloCfg.AppID = baseconfigsource.GetStringFallback(cfg, "apollo", "app_id")
	apolloCfg.Cluster = baseconfigsource.GetStringFallback(cfg, "apollo", "cluster")
	if apolloCfg.Cluster == "" {
		apolloCfg.Cluster = "default"
	}
	apolloCfg.Namespace = baseconfigsource.GetStringFallback(cfg, "apollo", "namespace")
	if apolloCfg.Namespace == "" {
		apolloCfg.Namespace = "application"
	}
	apolloCfg.MetaServer = baseconfigsource.GetStringFallback(cfg, "apollo", "meta_server")
	apolloCfg.AccessKey = baseconfigsource.GetStringFallback(cfg, "apollo", "access_key")

	if d := baseconfigsource.GetDurationSecondsFallback(cfg, "apollo", "poll_interval_seconds"); d > 0 {
		apolloCfg.PollInterval = d
	}
	if d := baseconfigsource.GetDurationMillisFallback(cfg, "apollo", "watch_retry_interval_ms"); d > 0 {
		apolloCfg.WatchRetryInterval = d
	}

	return apolloCfg, nil
}
