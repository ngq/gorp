package discovery

import (
	"context"
	"errors"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const (
	SelectorKey        = "framework.selector"
	SelectorBuilderKey = "framework.selector.builder"
)

type Selector interface {
	Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...SelectOption) (selected transportcontract.ServiceInstance, done DoneFunc, err error)
}

type SelectorBuilder interface {
	Build() Selector
}

type DoneFunc func(ctx context.Context, info DoneInfo)

type DoneInfo struct {
	Err           error
	ReplyMD       ReplyMetadata
	BytesSent     bool
	BytesReceived bool
}

type ReplyMetadata interface {
	Get(key string) string
}

type SelectOption func(*SelectOptions)

type SelectOptions struct {
	ForceInstance *transportcontract.ServiceInstance
	Filters       []NodeFilter
	Metadata      map[string]string
}

type NodeFilter func(instance transportcontract.ServiceInstance) bool

type WeightedNode interface {
	ServiceInstance() transportcontract.ServiceInstance
	Weight() float64
}

type SelectorAlgorithm string

const (
	SelectorNoop   SelectorAlgorithm = "noop"
	SelectorRandom SelectorAlgorithm = "random"
	SelectorWRR    SelectorAlgorithm = "wrr"
	SelectorP2C    SelectorAlgorithm = "p2c"
)

var ErrNoAvailable = errors.New("selector: no available service instance")

func WithForceInstance(instance transportcontract.ServiceInstance) SelectOption {
	return func(opts *SelectOptions) {
		opts.ForceInstance = &instance
	}
}

func WithFilter(filter NodeFilter) SelectOption {
	return func(opts *SelectOptions) {
		opts.Filters = append(opts.Filters, filter)
	}
}

func WithMetadata(key, value string) SelectOption {
	return func(opts *SelectOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
	}
}
