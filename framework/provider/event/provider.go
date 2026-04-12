package event

import (
	"github.com/ngq/gorp/framework/contract"
)

// Provider 事件服务提供者。
//
// 中文说明：
// - 将 LocalEventBus 注册到容器；
// - 后续可替换为 Redis/Kafka 等分布式实现；
// - 保持 contract.EventBus 接口不变。
type Provider struct{}

// NewProvider 创建事件服务提供者。
func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string      { return "event" }
func (p *Provider) IsDefer() bool     { return false }
func (p *Provider) Provides() []string { return []string{contract.EventKey} }

// Register 注册事件总线到容器。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.EventKey, func(c contract.Container) (interface{}, error) {
		return NewLocalEventBus(), nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }