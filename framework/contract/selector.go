package contract

import (
	"context"
	"errors"
)

const (
	// SelectorKey 是负载均衡选择器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于 RPC 客户端选择目标服务实例；
	// - noop 实现返回空选择器，单体项目零依赖；
	// - 微服务项目可启用 random/wrr/p2c 算法。
	SelectorKey = "framework.selector"

	// SelectorBuilderKey 是负载均衡构建器在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于创建 Selector 实例；
	// - 支持多种算法：noop/random/wrr/p2c。
	SelectorBuilderKey = "framework.selector.builder"
)

// Selector 定义负载均衡选择器接口。
//
// 中文说明：
// - 这是服务间调用的核心组件；
// - 从多个服务实例中选择一个目标；
// - 支持 random（随机）、wrr（加权轮询）、p2c（自适应）等算法；
// - noop 实现返回第一个实例或错误，单体项目零依赖。
//
// 设计说明：
// - instances 参数由调用方传入（从 Discovery 获取）；
// - 这样设计更简洁，Selector 不需要持有实例列表；
// - 与 Kratos 的 Pick(ctx, nodes) 模式类似。
type Selector interface {
	// Select 选择一个服务实例。
	//
	// 中文说明：
	// - ctx: 上下文，可携带元数据影响选择策略；
	// - instances: 可用实例列表（从 Discovery 获取）；
	// - opts: 可选参数，如强制指定某实例；
	// - 返回 selected（选中的实例）、done（调用完成回调）、err（错误）；
	// - done 必须在调用完成后执行，用于 P2C 等算法的权重调整。
	Select(ctx context.Context, instances []ServiceInstance, opts ...SelectOption) (selected ServiceInstance, done DoneFunc, err error)
}

// SelectorBuilder 构建负载均衡选择器。
//
// 中文说明：
// - 不同算法实现此接口；
// - Build() 返回具体的 Selector 实例。
type SelectorBuilder interface {
	// Build 构建 Selector 实例。
	Build() Selector
}

// DoneFunc 是调用完成后的回调函数。
//
// 中文说明：
// - 用于 P2C 等自适应算法的动态权重调整；
// - 必须在 RPC 调用完成后执行（无论成功或失败）；
// - done(ctx, DoneInfo) 记录调用结果，影响下次选择权重。
type DoneFunc func(ctx context.Context, info DoneInfo)

// DoneInfo 包含调用完成信息。
//
// 中文说明：
// - Err: 调用错误（如连接失败、超时）；
// - ReplyMD: 响应元数据（如服务端返回的权重提示）；
// - BytesSent: 是否已发送数据；
// - BytesReceived: 是否已接收响应。
type DoneInfo struct {
	// Err 调用错误
	Err error

	// ReplyMD 响应元数据
	ReplyMD ReplyMetadata

	// BytesSent 是否已发送数据
	BytesSent bool

	// BytesReceived 是否已接收响应
	BytesReceived bool
}

// ReplyMetadata 定义响应元数据接口。
//
// 中文说明：
// - 服务端可返回元数据影响客户端负载均衡；
// - 如返回 "lb-weight: 80" 提示客户端调整权重。
type ReplyMetadata interface {
	// Get 获取元数据值。
	Get(key string) string
}

// SelectOption 选择参数。
//
// 中文说明：
// - 用于 Select 方法的可选参数；
// - 可扩展支持：强制指定实例、过滤实例等。
type SelectOption func(*SelectOptions)

// SelectOptions 选择参数配置。
type SelectOptions struct {
	// ForceInstance 强制选择指定实例（忽略负载均衡）
	ForceInstance *ServiceInstance

	// Filters 实例过滤器
	Filters []NodeFilter

	// Metadata 上下文元数据
	Metadata map[string]string
}

// NodeFilter 实例过滤器。
//
// 中文说明：
// - 用于过滤不合适的实例；
// - 如过滤掉 unhealthy 实例、过滤特定版本等。
type NodeFilter func(instance ServiceInstance) bool

// WeightedNode 定义带权重的服务实例接口。
//
// 中文说明：
// - 用于加权负载均衡算法（WRR、P2C）；
// - 权重值可从实例元数据中提取；
// - 默认权重为 1.0。
type WeightedNode interface {
	// ServiceInstance 返回原始服务实例
	ServiceInstance() ServiceInstance

	// Weight 返回实例权重
	//
	// 中文说明：
	// - 权重范围通常为 0-100；
	// - 权重越高，被选中的概率越大；
	// - 元数据中可设置 "weight: 80" 自定义权重。
	Weight() float64
}

// SelectorAlgorithm 定义负载均衡算法类型。
//
// 中文说明：
// - noop: 单体项目零依赖，无负载均衡；
// - random: 随机选择，简单高效；
// - wrr: 加权轮询，根据权重分配；
// - p2c: 自适应负载均衡，根据延迟动态调整权重。
type SelectorAlgorithm string

const (
	// SelectorNoop noop 模式，单体项目零依赖
	SelectorNoop SelectorAlgorithm = "noop"

	// SelectorRandom 随机选择算法
	SelectorRandom SelectorAlgorithm = "random"

	// SelectorWRR 加权轮询算法
	SelectorWRR SelectorAlgorithm = "wrr"

	// SelectorP2C 自适应负载均衡算法
	SelectorP2C SelectorAlgorithm = "p2c"
)

// ErrNoAvailable 表示无可用服务实例。
//
// 中文说明：
// - 当所有实例都不健康或实例列表为空时返回此错误；
// - 微服务项目应配置服务发现以避免此错误；
// - 单体项目使用 noop 模式不应触发此错误。
var ErrNoAvailable = errors.New("selector: no available service instance")

// WithForceInstance 强制选择指定实例。
//
// 中文说明：
// - 跳过负载均衡，直接使用指定实例；
// - 用于调试、熔断恢复等场景。
func WithForceInstance(instance ServiceInstance) SelectOption {
	return func(opts *SelectOptions) {
		opts.ForceInstance = &instance
	}
}

// WithFilter 添加实例过滤器。
//
// 中文说明：
// - 用于过滤不合适的实例；
// - 如过滤 unhealthy 实例。
func WithFilter(filter NodeFilter) SelectOption {
	return func(opts *SelectOptions) {
		opts.Filters = append(opts.Filters, filter)
	}
}

// WithMetadata 添加上下文元数据。
//
// 中文说明：
// - 用于传递影响负载均衡的元数据；
// - 如 "version: v2" 优先选择 v2 版本实例。
func WithMetadata(key, value string) SelectOption {
	return func(opts *SelectOptions) {
		if opts.Metadata == nil {
			opts.Metadata = make(map[string]string)
		}
		opts.Metadata[key] = value
	}
}