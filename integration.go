package gorp

import (
	"github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/contract/integration"
	"github.com/ngq/gorp/framework/application"
)

// DistributedLock is the top-level alias of the distributed lock contract.
// DistributedLock 是分布式锁契约的顶层别名。
type DistributedLock = data.DistributedLock

// MessagePublisher is the top-level alias of the message publisher contract.
// MessagePublisher 是消息发布契约的顶层别名。
type MessagePublisher = integration.MessagePublisher

// MessageSubscriber is the top-level alias of the message subscriber contract.
// MessageSubscriber 是消息订阅契约的顶层别名。
type MessageSubscriber = integration.MessageSubscriber

// Message is the top-level alias of the integration message contract.
// Message 是集成消息契约的顶层别名。
type Message = integration.Message

// MakeDistributedLock returns the distributed lock capability from the container.
// MakeDistributedLock 从容器中获取分布式锁能力。
func MakeDistributedLock(c Container) (DistributedLock, error) {
	return application.MakeDistributedLock(c)
}

// MakeMessagePublisher returns the message publishing capability from the container.
// MakeMessagePublisher 从容器中获取消息发布能力。
func MakeMessagePublisher(c Container) (MessagePublisher, error) {
	return application.MakeMessagePublisher(c)
}

// MakeMessageSubscriber returns the message subscription capability from the container.
// MakeMessageSubscriber 从容器中获取消息订阅能力。
func MakeMessageSubscriber(c Container) (MessageSubscriber, error) {
	return application.MakeMessageSubscriber(c)
}
