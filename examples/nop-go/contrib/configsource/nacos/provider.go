// Package nacos provides Nacos configuration center provider for the gorp framework.
// This provider implements ConfigSource contract with Nacos SDK integration.
//
// 本包提供 gorp 框架 Nacos 配置中心 provider。
// 本 provider 实现 ConfigSource 契约，集成 Nacos SDK。
package nacos

import (
	"time"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider 提供 Nacos 配置中心实现。
type Provider struct {
	BaseConfigSourceProvider
}

// NewProvider creates a new Nacos provider instance.
func NewProvider() *Provider {
	p := &Provider{}
	p.NameStr = "configsource.nacos"
	p.GetConfig = func(c runtimecontract.Container) (any, error) {
		return getNacosConfig(c)
	}
	p.NewSource = func(cfg any) (datacontract.ConfigSource, error) {
		return NewConfigSource(cfg.(*NacosConfig))
	}
	return p
}

// NacosConfig 定义 Nacos 配置。
type NacosConfig struct {
	ServerAddr   string
	Port         int
	Namespace    string
	Group        string
	DataID       string
	Username     string
	Password     string
	PollInterval time.Duration
}

// getNacosConfig extracts Nacos configuration from the container's config binding.
// Uses GetStringFallback for single-path reading, eliminating the cfg.Get()+cfg.GetString() double-read.
//
// getNacosConfig 从容器的 config binding 中提取 Nacos 配置。
// 使用 GetStringFallback 单路径读取，消除 cfg.Get()+cfg.GetString() 双读冗余。
func getNacosConfig(c runtimecontract.Container) (*NacosConfig, error) {
	cfg, err := ReadConfig(c)
	if err != nil {
		return nil, err
	}

	nacosCfg := &NacosConfig{
		Port:         8848,
		Group:        "DEFAULT_GROUP",
		PollInterval: defaultNacosPollInterval,
	}

	nacosCfg.ServerAddr = GetStringFallback(cfg, "nacos", "server_addr")
	if port := GetIntFallback(cfg, "nacos", "port"); port > 0 {
		nacosCfg.Port = port
	}
	nacosCfg.Namespace = GetStringFallback(cfg, "nacos", "namespace")
	nacosCfg.Group = GetStringFallback(cfg, "nacos", "group")
	if nacosCfg.Group == "" {
		nacosCfg.Group = "DEFAULT_GROUP"
	}
	nacosCfg.DataID = GetStringFallback(cfg, "nacos", "data_id")
	nacosCfg.Username = GetStringFallback(cfg, "nacos", "username")
	nacosCfg.Password = GetStringFallback(cfg, "nacos", "password")

	if d := GetDurationSecondsFallback(cfg, "nacos", "poll_interval_seconds"); d > 0 {
		nacosCfg.PollInterval = d
	}

	return nacosCfg, nil
}
