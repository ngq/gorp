package gorp

import (
	"github.com/ngq/gorp/framework/contract/data"
	"github.com/ngq/gorp/framework/contract/integration"
	"github.com/ngq/gorp/framework/facade"
)

type DistributedLock = data.DistributedLock
type MessagePublisher = integration.MessagePublisher
type MessageSubscriber = integration.MessageSubscriber
type Message = integration.Message

func MakeDistributedLock(c Container) (DistributedLock, error) {
	return facade.MakeDistributedLock(c)
}

func MakeMessagePublisher(c Container) (MessagePublisher, error) {
	return facade.MakeMessagePublisher(c)
}

func MakeMessageSubscriber(c Container) (MessageSubscriber, error) {
	return facade.MakeMessageSubscriber(c)
}
