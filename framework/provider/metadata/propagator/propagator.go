package propagator

import (
	"context"
	"strings"

	"github.com/ngq/gorp/framework/contract"
)

// DefaultPropagator 默认 metadata 传播器。
//
// 中文说明：
// - 实现服务间 metadata 自动传递；
// - 支持前缀过滤（默认 x-md- 前缀）；
// - 支持常量 metadata（每次请求都携带）；
// - 自己实现，不抄袭 Kratos。
type DefaultPropagator struct {
	// propagatePrefix 需要传播的 key 前缀
	propagatePrefix []string

	// constantMetadata 常量 metadata
	constantMetadata map[string]string
}

// NewDefaultPropagator 创建默认传播器。
//
// 中文说明：
// - prefix: 需要传播的 key 前缀（如 "x-md-"）；
// - constant: 常量 metadata（每次请求都携带）。
func NewDefaultPropagator(prefix []string, constant map[string]string) *DefaultPropagator {
	if len(prefix) == 0 {
		prefix = []string{"x-md-"}
	}
	if constant == nil {
		constant = make(map[string]string)
	}
	return &DefaultPropagator{
		propagatePrefix:  prefix,
		constantMetadata: constant,
	}
}

// Inject 从 context 提取 metadata 注入 carrier。
//
// 中文说明：
// - 服务端 -> 客户端 调用时使用；
// - 从 server context 读取 metadata；
// - 注入到 HTTP Header / gRPC Metadata；
// - 只注入匹配前缀的 key。
func (p *DefaultPropagator) Inject(ctx context.Context, carrier contract.MetadataCarrier) {
	// 1. 注入常量 metadata
	for k, v := range p.constantMetadata {
		carrier.Set(k, v)
	}

	// 2. 注入客户端 metadata
	if clientMD, ok := contract.FromClientContext(ctx); ok {
		clientMD.Range(func(key string, values []string) bool {
			for _, v := range values {
				carrier.Add(key, v)
			}
			return true
		})
	}

	// 3. 注入服务端 metadata（只传播匹配前缀的）
	if serverMD, ok := contract.FromServerContext(ctx); ok {
		serverMD.Range(func(key string, values []string) bool {
			if p.matchPrefix(key) {
				for _, v := range values {
					carrier.Add(key, v)
				}
			}
			return true
		})
	}
}

// Extract 从 carrier 提取 metadata 存入 context。
//
// 中文说明：
// - 客户端 -> 服务端 调用时使用；
// - 从 HTTP Header / gRPC Metadata 提取；
// - 存入 server context；
// - 只提取匹配前缀的 key。
func (p *DefaultPropagator) Extract(ctx context.Context, carrier contract.MetadataCarrier) context.Context {
	// 创建新的 metadata
	md := contract.NewMetadata()

	// 遍历 carrier 的所有 key
	for _, key := range carrier.Keys() {
		// 只提取匹配前缀的 key
		if p.matchPrefix(key) {
			values := carrier.Values(key)
			for _, v := range values {
				md.Add(key, v)
			}
		}
	}

	// 添加常量 metadata
	for k, v := range p.constantMetadata {
		md.Set(k, v)
	}

	// 存入 context
	return contract.NewServerContext(ctx, md)
}

// matchPrefix 检查 key 是否匹配任一前缀。
//
// 中文说明：
// - key 转为小写后匹配；
// - 前缀列表为空时匹配所有 key。
func (p *DefaultPropagator) matchPrefix(key string) bool {
	if len(p.propagatePrefix) == 0 {
		return true
	}
	lowerKey := strings.ToLower(key)
	for _, prefix := range p.propagatePrefix {
		if strings.HasPrefix(lowerKey, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}

// NoopPropagator 空传播器。
//
// 中文说明：
// - 单体项目使用；
// - 不传播任何 metadata。
type NoopPropagator struct{}

func NewNoopPropagator() *NoopPropagator { return &NoopPropagator{} }

func (p *NoopPropagator) Inject(ctx context.Context, carrier contract.MetadataCarrier) {
	// 空操作
}

func (p *NoopPropagator) Extract(ctx context.Context, carrier contract.MetadataCarrier) context.Context {
	return ctx
}