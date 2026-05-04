package noop

import (
	"context"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// Provider 提供 noop metadata 实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖；
// - 所有操作返回空值或空操作，单体项目无需 metadata 传递。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "metadata.noop" }
func (p *Provider) IsDefer() bool { return true }
func (p *Provider) Provides() []string {
	return []string{transportcontract.MetadataKey, transportcontract.MetadataPropagatorKey}
}

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(transportcontract.MetadataKey, func(c runtimecontract.Container) (any, error) {
		return &noopMetadata{}, nil
	}, true)
	c.Bind(transportcontract.MetadataPropagatorKey, func(c runtimecontract.Container) (any, error) {
		return &noopPropagator{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(c runtimecontract.Container) error {
	return nil
}

// noopMetadata 是 Metadata 的空实现。
//
// 中文说明：
// - Get 返回空字符串；
// - Values 返回空切片；
// - Set/Add/Del 空操作；
// - 单体项目无需 metadata 传递。
type noopMetadata struct{}

func (m *noopMetadata) Get(key string) string               { return "" }
func (m *noopMetadata) Values(key string) []string          { return nil }
func (m *noopMetadata) Set(key, value string)               {}
func (m *noopMetadata) Add(key, value string)               {}
func (m *noopMetadata) Del(key string)                      {}
func (m *noopMetadata) Range(f func(string, []string) bool) {}
func (m *noopMetadata) Clone() transportcontract.Metadata   { return &noopMetadata{} }
func (m *noopMetadata) ToMap() map[string][]string          { return nil }

// noopPropagator 是 Propagator 的空实现。
type noopPropagator struct{}

func (p *noopPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
	// 空操作
}

func (p *noopPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	return ctx
}
