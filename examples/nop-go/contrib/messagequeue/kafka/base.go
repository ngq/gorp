// Package kafka 提供 gorp 框架基于 Kafka 的消息队列 provider。
// 本文件内联了 basemq.BaseMQProvider 和 native.As，消除对 contrib/internal 的依赖，
// 使本包成为可独立引用的模块。
package kafka

import (
	"fmt"
	"reflect"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// As 尝试通过 reflect 将 source 转换为 target 指向的类型。
// 支持直接赋值、接口实现和类型转换三种路径。
// 当 target 为 nil、非指针、nil 指针或不可设置时返回 false。
func As(source any, target any) bool {
	if target == nil {
		return false
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return false
	}
	sourceValue := reflect.ValueOf(source)
	if !sourceValue.IsValid() {
		return false
	}
	destination := targetValue.Elem()
	if !destination.CanSet() {
		return false
	}
	sourceType := sourceValue.Type()
	destinationType := destination.Type()
	if sourceType.AssignableTo(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if destinationType.Kind() == reflect.Interface && sourceType.Implements(destinationType) {
		destination.Set(sourceValue)
		return true
	}
	if sourceType.ConvertibleTo(destinationType) {
		destination.Set(sourceValue.Convert(destinationType))
		return true
	}
	return false
}

// BaseMQProvider 消除各 MQ provider 之间的结构重复。
// 内联自 contrib/internal/basemq，使本包成为独立模块。
//
// 各字段含义：
//   - NameStr: provider 名称标识
//   - GetConfig: 从容器获取 MQ 配置的回调
//   - NewQueue: 根据配置创建 MQ 实例的回调
type BaseMQProvider struct {
	NameStr   string
	GetConfig func(c runtimecontract.Container) (*integrationcontract.MessageQueueConfig, error)
	NewQueue  func(cfg *integrationcontract.MessageQueueConfig) (integrationcontract.MessageQueue, error)
}

// Name 返回 provider 名称标识。
func (p *BaseMQProvider) Name() string { return p.NameStr }

// IsDefer 返回 true，表示 provider 延迟初始化。
func (p *BaseMQProvider) IsDefer() bool { return true }

// Provides 返回该 provider 提供的契约键列表：
// MessageQueueKey、MessagePublisherKey、MessageSubscriberKey。
func (p *BaseMQProvider) Provides() []string {
	return []string{integrationcontract.MessageQueueKey, integrationcontract.MessagePublisherKey, integrationcontract.MessageSubscriberKey}
}

// DependsOn 返回该 provider 依赖的契约键列表：ConfigKey。
func (p *BaseMQProvider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Boot 延迟 provider 无需 boot 操作，直接返回 nil。
func (p *BaseMQProvider) Boot(runtimecontract.Container) error { return nil }

// Register 将 MessageQueue、MessagePublisher、MessageSubscriber 绑定到容器。
// 创建 MQ 实例后会注册 closer，在容器销毁时自动关闭资源。
func (p *BaseMQProvider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.MessageQueueKey, func(c runtimecontract.Container) (any, error) {
		cfg, err := p.GetConfig(c)
		if err != nil {
			return nil, err
		}
		q, err := p.NewQueue(cfg)
		if err != nil {
			return nil, err
		}
		c.RegisterCloser(integrationcontract.MessageQueueKey, q)
		return q, nil
	}, true)
	c.Bind(integrationcontract.MessagePublisherKey, func(c runtimecontract.Container) (any, error) {
		mq, err := c.Make(integrationcontract.MessageQueueKey)
		if err != nil {
			return nil, err
		}
		mqSvc, ok := mq.(integrationcontract.MessageQueue)
		if !ok {
			return nil, fmt.Errorf("basemq: expected integrationcontract.MessageQueue, got %T", mq)
		}
		return mqSvc.Publisher(), nil
	}, true)
	c.Bind(integrationcontract.MessageSubscriberKey, func(c runtimecontract.Container) (any, error) {
		mq, err := c.Make(integrationcontract.MessageQueueKey)
		if err != nil {
			return nil, err
		}
		mqSvc, ok := mq.(integrationcontract.MessageQueue)
		if !ok {
			return nil, fmt.Errorf("basemq: expected integrationcontract.MessageQueue, got %T", mq)
		}
		return mqSvc.Subscriber(), nil
	}, true)
	return nil
}
