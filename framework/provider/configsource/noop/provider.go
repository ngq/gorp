package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 配置源实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何远程配置依赖（无 Consul/etcd/Nacos）；
// - 所有方法返回空值或错误，本地配置由 config provider 负责。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.ConfigSourceKey, contract.ConfigWatcherKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		return &noopSource{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrNoopConfigSource 表示 noop 配置源不支持远程操作。
var ErrNoopConfigSource = errors.New("configsource: noop mode, remote config not available in monolith")

// noopSource 是 ConfigSource 的空实现。
//
// 中文说明：
// - 单体项目使用本地文件配置，不需要远程配置源；
// - 所有方法返回错误或空值。
type noopSource struct{}

func (s *noopSource) Load(ctx context.Context) (map[string]any, error) {
	// 返回空配置，本地配置由 config provider 处理
	return map[string]any{}, nil
}

func (s *noopSource) Get(ctx context.Context, key string) (any, error) {
	// noop 不存储配置，返回不支持错误
	return nil, ErrNoopConfigSource
}

func (s *noopSource) Set(ctx context.Context, key string, value any) error {
	return ErrNoopConfigSource
}

func (s *noopSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return nil, ErrNoopConfigSource
}

func (s *noopSource) Close() error { return nil }