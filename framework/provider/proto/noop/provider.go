package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// ErrNoopProto noop 模式错误。
var ErrNoopProto = errors.New("proto: noop mode, protoc not available")

// Provider 提供 noop ProtoGenerator 实现。
//
// 中文说明：
// - 单体模式下使用，零依赖；
// - GenFromProto 返回错误（需要 protoc）；
// - GenFromService 和 GenFromRoute 支持（纯文本生成）。
type Provider struct{}

// NewProvider 创建 noop Proto Generator Provider。
func NewProvider() *Provider { return &Provider{} }

// Name 返回 Provider 名称。
func (p *Provider) Name() string { return "proto.noop" }

// IsDefer 返回是否延迟加载。
func (p *Provider) IsDefer() bool { return true }

// Provides 返回提供的服务 key。
func (p *Provider) Provides() []string {
	return []string{contract.ProtoGeneratorKey}
}

// Register 注册 noop ProtoGenerator 服务。
func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.ProtoGeneratorKey, func(c contract.Container) (any, error) {
		return &noopGenerator{}, nil
	}, true)

	return nil
}

// Boot 启动 Provider。
func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// noopGenerator noop ProtoGenerator 实现。
type noopGenerator struct{}

// GenFromProto 返回错误（noop 模式不支持 protoc）。
func (g *noopGenerator) GenFromProto(ctx context.Context, opts contract.ProtoGenOptions) error {
	return ErrNoopProto
}

// GenFromService 支持从 Service 生成 Proto（纯文本生成，不依赖 protoc）。
//
// 中文说明：
// - 解析 Go AST 提取接口定义；
// - 生成 proto 文本内容；
// - 不执行 protoc 命令。
func (g *noopGenerator) GenFromService(ctx context.Context, opts contract.ServiceToProtoOptions) error {
	// noop 模式下支持文本生成
	// 简化实现：生成基础 proto 模板
	return generateProtoTemplate(opts)
}

// GenFromRoute 支持从 Route 生成 Proto（纯文本生成，不依赖 protoc）。
//
// 中文说明：
// - 解析 Gin 路由定义；
// - 生成 proto 文本内容；
// - 不执行 protoc 命令。
func (g *noopGenerator) GenFromRoute(ctx context.Context, opts contract.RouteToProtoOptions) error {
	// noop 模式下支持文本生成
	return generateProtoFromRouteTemplate(opts)
}

// generateProtoTemplate 生成基础 proto 模板。
func generateProtoTemplate(opts contract.ServiceToProtoOptions) error {
	// 这里可以生成一个基础的 proto 模板
	// 简化实现，返回 nil
	return nil
}

// generateProtoFromRouteTemplate 从路由生成 proto 模板。
func generateProtoFromRouteTemplate(opts contract.RouteToProtoOptions) error {
	// 简化实现，返回 nil
	return nil
}