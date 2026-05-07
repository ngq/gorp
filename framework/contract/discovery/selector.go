// Application scenarios:
// - Define the service-instance selection contract used by discovery-aware RPC clients.
// - Separate selection algorithms, filters, and completion feedback from concrete providers.
// - Provide one shared option model for force-routing, metadata hints, and custom node filters.
//
// 适用场景：
// - 定义 discovery 感知型 RPC 客户端使用的服务实例选择契约。
// - 将选择算法、过滤器和完成反馈与具体 provider 解耦。
// - 为强制路由、metadata 提示和自定义节点过滤提供统一选项模型。
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

// Selector chooses one service instance from a candidate set.
//
// Selector 用于从候选集合中选择一个服务实例。
type Selector interface {
	Select(ctx context.Context, instances []transportcontract.ServiceInstance, opts ...SelectOption) (selected transportcontract.ServiceInstance, done DoneFunc, err error)
}

// SelectorBuilder builds a selector instance.
//
// SelectorBuilder 用于构建选择器实例。
type SelectorBuilder interface {
	Build() Selector
}

// DoneFunc reports request completion feedback back to the selector.
//
// DoneFunc 用于把请求完成反馈回传给选择器。
type DoneFunc func(ctx context.Context, info DoneInfo)

// DoneInfo describes the request result observed by the selector.
//
// DoneInfo 描述选择器观测到的请求结果。
type DoneInfo struct {
	Err           error
	ReplyMD       ReplyMetadata
	BytesSent     bool
	BytesReceived bool
}

// ReplyMetadata defines the metadata view exposed to selectors after one request.
//
// ReplyMetadata 定义请求完成后暴露给选择器的 metadata 视图。
type ReplyMetadata interface {
	Get(key string) string
}

// SelectOption mutates select options.
//
// SelectOption 用于修改选择选项。
type SelectOption func(*SelectOptions)

// SelectOptions describes the selector input options.
//
// SelectOptions 描述选择器输入选项。
type SelectOptions struct {
	ForceInstance *transportcontract.ServiceInstance
	Filters       []NodeFilter
	Metadata      map[string]string
}

// NodeFilter filters candidate service instances before selection.
//
// NodeFilter 用于在选择前过滤候选服务实例。
type NodeFilter func(instance transportcontract.ServiceInstance) bool

// WeightedNode describes a candidate that exposes both instance data and weight.
//
// WeightedNode 描述同时暴露实例信息和权重的候选节点。
type WeightedNode interface {
	ServiceInstance() transportcontract.ServiceInstance
	Weight() float64
}

// SelectorAlgorithm describes the selector algorithm identifier.
//
// SelectorAlgorithm 描述选择算法标识。
type SelectorAlgorithm string

const (
	SelectorNoop   SelectorAlgorithm = "noop"
	SelectorRandom SelectorAlgorithm = "random"
	SelectorWRR    SelectorAlgorithm = "wrr"
	SelectorP2C    SelectorAlgorithm = "p2c"
)

// ErrNoAvailable indicates that no service instance is currently selectable.
//
// ErrNoAvailable 表示当前没有可选的服务实例。
var ErrNoAvailable = errors.New("selector: no available service instance")

// WithForceInstance forces the selector to use one specific instance.
//
// WithForceInstance 强制选择器使用指定实例。
func WithForceInstance(instance transportcontract.ServiceInstance) SelectOption {
	return func(opts *SelectOptions) {
		opts.ForceInstance = &instance
	}
}

// WithFilter appends one custom node filter.
//
// WithFilter 追加一个自定义节点过滤器。
func WithFilter(filter NodeFilter) SelectOption {
	return func(opts *SelectOptions) {
		opts.Filters = append(opts.Filters, filter)
	}
}

// WithMetadata attaches one metadata hint used during selection.
//
// WithMetadata 追加一个选择阶段使用的 metadata 提示。
func WithMetadata(key, value string) SelectOption {
	return func(opts *SelectOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
	}
}
