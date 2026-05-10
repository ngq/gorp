// Package propagator provides metadata propagation implementation for gorp framework.
// Supports prefix-based filtering and constant metadata injection.
// Used for HTTP/gRPC cross-boundary metadata propagation.
//
// 传播器包提供元数据传播实现，用于 gorp 框架。
// 支持前缀过滤和常量元数据注入。
// 用于 HTTP/gRPC 跨边界元数据传播。
package propagator

import (
	"context"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// DefaultPropagator implements MetadataPropagator with prefix filtering.
// Core logic: Filter keys by prefix, inject constant metadata, propagate client/server metadata.
//
// DefaultPropagator 实现带前缀过滤的 MetadataPropagator。
// 核心逻辑：按前缀过滤键、注入常量元数据、传播客户端/服务端元数据。
type DefaultPropagator struct {
	propagatePrefix  []string
	constantMetadata map[string]string
}

// NewDefaultPropagator creates a propagator with prefix and constant metadata.
// Core logic: Set default prefix if empty, initialize constant map.
//
// NewDefaultPropagator 创建带前缀和常量元数据的传播器。
// 核心逻辑：若前缀为空则设置默认值、初始化常量 map。
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

// Inject injects metadata into carrier for outgoing request.
// Core logic: Inject constant metadata, then client/server metadata matching prefix.
//
// Inject 将元数据注入 carrier，用于 outgoing 请求。
// 核心逻辑：注入常量元数据、然后注入匹配前缀的客户端/服务端元数据。
func (p *DefaultPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
	for k, v := range p.constantMetadata {
		carrier.Set(k, v)
	}

	if clientMD, ok := transportcontract.FromClientContext(ctx); ok {
		clientMD.Range(func(key string, values []string) bool {
			for _, v := range values {
				carrier.Add(key, v)
			}
			return true
		})
	}

	if serverMD, ok := transportcontract.FromServerContext(ctx); ok {
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

// Extract extracts metadata from carrier into context for incoming request.
// Core logic: Extract keys matching prefix, create server context with metadata.
//
// Extract 从 carrier 提取元数据到 context，用于 incoming 请求。
// 核心逻辑：提取匹配前缀的键、创建携带元数据的 server context。
func (p *DefaultPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	md := transportcontract.NewMetadata()

	for _, key := range carrier.Keys() {
		if p.matchPrefix(key) {
			values := carrier.Values(key)
			for _, v := range values {
				md.Add(key, v)
			}
		}
	}

	for k, v := range p.constantMetadata {
		md.Set(k, v)
	}

	return transportcontract.NewServerContext(ctx, md)
}

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

// NoopPropagator is a no-op implementation of MetadataPropagator.
// Used when metadata propagation is disabled.
//
// NoopPropagator 是 MetadataPropagator 的空实现。
// 用于禁用元数据传播的场景。
type NoopPropagator struct{}

// NewNoopPropagator creates a no-op propagator.
//
// NewNoopPropagator 创建空传播器。
func NewNoopPropagator() *NoopPropagator { return &NoopPropagator{} }

// Inject does nothing (no-op).
//
// Inject 无操作。
func (p *NoopPropagator) Inject(ctx context.Context, carrier transportcontract.MetadataCarrier) {
}

// Extract returns original context unchanged.
//
// Extract 返回原始 context 不做修改。
func (p *NoopPropagator) Extract(ctx context.Context, carrier transportcontract.MetadataCarrier) context.Context {
	return ctx
}
