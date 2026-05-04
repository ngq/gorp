package noop

import (
	"context"
	"errors"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "configsource.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{datacontract.ConfigSourceKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.ConfigSourceKey, func(c runtimecontract.Container) (any, error) {
		return &noopConfigSource{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

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

func (s *noopConfigSource) Watch(_ context.Context, _ string) (datacontract.ConfigWatcher, error) {
	return nil, errors.New("configsource.noop: Watch not supported")
}

func (s *noopConfigSource) Close() error { return nil }
