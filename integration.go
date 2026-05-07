// Application scenarios:
// - Expose root-package aliases and helpers for integration capabilities such as message queue and distributed lock.
// - Keep common integration capability access on a short public path for business code.
// - Re-export stable integration primitives without requiring direct imports of contract packages.
//
// 适用场景：
// - 暴露根包层的集成能力别名和 helper，例如消息队列与分布式锁。
// - 让业务代码通过简短公共路径获取常用 integration 能力。
// - 在不要求业务直接导入 contract 包的前提下重导出稳定集成原语。
package gorp

import (
	"github.com/ngq/gorp/framework/application"
	"github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/contract/integration"
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
