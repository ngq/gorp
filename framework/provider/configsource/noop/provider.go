package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 配置源实现。
//
// 中文说明：
// - 单服务 / 本地场景默认使用此 provider；
// - 不从任何远程配置源加载，Load 返回空 map；
// - 需要远程配置中心时，注册 contrib/configsource/* 中的实现替换本 provider。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string      { return "configsource.noop" }
func (p *Provider) IsDefer() bool     { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		return &noopConfigSource{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

// noopConfigSource 是 contract.ConfigSource 的空实现。
type noopConfigSource struct{}

func (s *noopConfigSource) Load(_ context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}

func (s *noopConfigSource) Get(_ context.Context, _ string) (any, error) {
	return nil, nil
}

func (s *noopConfigSource) Set(_ context.Context, _ string, _ any) error {
	return errors.New("configsource.noop: Set not supported")
}

func (s *noopConfigSource) Watch(_ context.Context, _ string) (contract.ConfigWatcher, error) {
	return nil, errors.New("configsource.noop: Watch not supported")
}

func (s *noopConfigSource) Close() error { return nil }
