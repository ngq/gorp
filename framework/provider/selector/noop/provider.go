package noop

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop 负载均衡选择器实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖；
// - Select 返回空实例或错误，单体项目无需负载均衡；
// - 单体项目应直接调用本地服务，不应触发服务间调用。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "selector.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.SelectorKey, contract.SelectorBuilderKey}
}

func (p *Provider) Register(c contract.Container) error {
	// 注册 Selector Builder
	c.Bind(contract.SelectorBuilderKey, func(c contract.Container) (any, error) {
		return &noopBuilder{}, nil
	}, true)

	// 注册默认 Selector 实例
	c.Bind(contract.SelectorKey, func(c contract.Container) (any, error) {
		builder := &noopBuilder{}
		return builder.Build(), nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopBuilder 构建 noop Selector。
type noopBuilder struct{}

func (b *noopBuilder) Build() contract.Selector {
	return &noopSelector{}
}

// noopSelector 是 Selector 的空实现。
//
// 中文说明：
// - Select 返回 ErrNoAvailable，表示无远程服务；
// - 单体项目应使用本地调用，不应依赖负载均衡；
// - DoneFunc 是空操作，无需权重调整。
type noopSelector struct{}

// Select 选择服务实例（返回错误）。
//
// 中文说明：
// - 单体项目无远程服务，返回 ErrNoAvailable；
// - 如果 opts.ForceInstance 指定了实例，则返回该实例；
// - 如果 instances 不为空，返回第一个健康实例；
// - done 是空回调，无权重调整逻辑。
func (s *noopSelector) Select(ctx context.Context, instances []contract.ServiceInstance, opts ...contract.SelectOption) (
	selected contract.ServiceInstance, done contract.DoneFunc, err error,
) {
	// 解析可选参数
	options := &contract.SelectOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// 如果强制指定实例，则返回该实例
	if options.ForceInstance != nil {
		return *options.ForceInstance, noopDone, nil
	}

	// 如果实例列表不为空，返回第一个健康实例
	for _, instance := range instances {
		if instance.Healthy {
			return instance, noopDone, nil
		}
	}

	// 单体项目无远程服务，返回错误
	return contract.ServiceInstance{}, noopDone, contract.ErrNoAvailable
}

// noopDone 是空的 DoneFunc 实现。
//
// 中文说明：
// - 无权重调整逻辑；
// - 无性能统计；
// - 单体项目无需调用完成回调。
func noopDone(ctx context.Context, info contract.DoneInfo) {
	// 空操作
}