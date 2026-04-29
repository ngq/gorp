package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "configsource.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string { return []string{contract.ConfigSourceKey, contract.ConfigWatcherKey} }

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ConfigSourceKey, func(c contract.Container) (any, error) {
		return &noopSource{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(c contract.Container) error { return nil }

var ErrNoopConfigSource = errors.New("configsource: noop mode, remote config not available in monolith")

type noopSource struct{}

func (s *noopSource) Load(ctx context.Context) (map[string]any, error) { return map[string]any{}, nil }
func (s *noopSource) Get(ctx context.Context, key string) (any, error) { return nil, ErrNoopConfigSource }
func (s *noopSource) Set(ctx context.Context, key string, value any) error { return ErrNoopConfigSource }
func (s *noopSource) Watch(ctx context.Context, key string) (contract.ConfigWatcher, error) {
	return nil, ErrNoopConfigSource
}
func (s *noopSource) Close() error { return nil }
