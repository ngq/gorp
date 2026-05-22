// Package noop provides a no-op DTM distributed transaction client for monolith scenarios.
// This client returns ErrNoopDTM for all transaction operations.
// Note: Distributed transactions are not available in monolith mode.
//
// 空 DTM 分布式事务客户端实现包，用于单体应用场景。
// 此客户端对所有事务操作返回 ErrNoopDTM 错误。
// 注意：分布式事务在单体模式下不可用。
package noop

import (
	"context"
	"errors"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers a no-op DTM client contract.
//
// Provider 注册空 DTM 客户端契约。
type Provider struct{}

// NewProvider creates a new no-op DTM provider instance.
//
// NewProvider 创建新的空 DTM Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "dtm.noop".
//
// Name 返回 Provider 名称 "dtm.noop"。
func (p *Provider) Name() string { return "dtm.noop" }

// IsDefer returns true, DTM client can be deferred until first use.
//
// IsDefer 返回 true，DTM 客户端可延迟初始化直到首次使用。
func (p *Provider) IsDefer() bool { return true }

// Provides returns the DTM client contract key.
//
// Provides 返回 DTM 客户端契约键。
func (p *Provider) Provides() []string { return []string{integrationcontract.DTMKey} }

// DependsOn returns the keys this provider depends on.
// Noop DTM has no dependencies.
//
// DependsOn 返回该 provider 依赖的 key。
// Noop DTM 无依赖。
func (p *Provider) DependsOn() []string { return nil }

// Register binds the no-op DTM client to the container.
//
// Register 将空 DTM 客户端绑定到容器。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.DTMKey, func(c runtimecontract.Container) (any, error) {
		return &noopDTMClient{}, nil
	}, true)
	return nil
}

// Boot is a no-op for this provider.
//
// Boot 此 Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }

// ErrNoopDTM indicates DTM is not available in monolith mode.
//
// ErrNoopDTM 表示 DTM 在单体模式下不可用。
var ErrNoopDTM = errors.New("dtm: noop mode, distributed transaction not available in monolith")

// noopDTMClient implements integrationcontract.DTMClient with no-op behavior.
//
// noopDTMClient 使用空行为实现 integrationcontract.DTMClient 接口。
type noopDTMClient struct{}

// SAGA returns a no-op SAGA builder.
//
// SAGA 返回空 SAGA 构建器。
func (c *noopDTMClient) SAGA(name string) integrationcontract.SAGABuilder {
	_ = name
	return &noopSAGABuilder{}
}

// TCC returns a no-op TCC builder.
//
// TCC 返回空 TCC 构建器。
func (c *noopDTMClient) TCC(name string) integrationcontract.TCCBuilder {
	_ = name
	return &noopTCCBuilder{}
}

// XA returns a no-op XA builder.
//
// XA 返回空 XA 构建器。
func (c *noopDTMClient) XA(name string) integrationcontract.XABuilder {
	_ = name
	return &noopXABuilder{}
}

// Barrier returns a no-op barrier handler.
//
// Barrier 返回空屏障处理器。
func (c *noopDTMClient) Barrier(transType, gid string) integrationcontract.BarrierHandler {
	_ = transType
	_ = gid
	return &noopBarrierHandler{}
}

// Query returns ErrNoopDTM.
//
// Query 返回 ErrNoopDTM。
func (c *noopDTMClient) Query(ctx context.Context, gid string) (*integrationcontract.TransactionInfo, error) {
	_ = ctx
	_ = gid
	return nil, ErrNoopDTM
}

// noopSAGABuilder implements SAGABuilder with no-op behavior.
//
// noopSAGABuilder 使用空行为实现 SAGABuilder 接口。
type noopSAGABuilder struct{}

func (b *noopSAGABuilder) Add(action string, compensate string, payload any) integrationcontract.SAGABuilder {
	_, _, _ = action, compensate, payload
	return b
}

func (b *noopSAGABuilder) AddBranch(action string, compensate string, payload any, opts integrationcontract.BranchOptions) integrationcontract.SAGABuilder {
	_, _, _, _ = action, compensate, payload, opts
	return b
}

func (b *noopSAGABuilder) Submit(ctx context.Context) error {
	_ = ctx
	return ErrNoopDTM
}

func (b *noopSAGABuilder) Build() (*integrationcontract.SAGATransaction, error) {
	return nil, ErrNoopDTM
}

// noopTCCBuilder implements TCCBuilder with no-op behavior.
//
// noopTCCBuilder 使用空行为实现 TCCBuilder 接口。
type noopTCCBuilder struct{}

func (b *noopTCCBuilder) Add(try string, confirm string, cancel string, payload any) integrationcontract.TCCBuilder {
	_, _, _, _ = try, confirm, cancel, payload
	return b
}

func (b *noopTCCBuilder) Submit(ctx context.Context) error {
	_ = ctx
	return ErrNoopDTM
}

// noopXABuilder implements XABuilder with no-op behavior.
//
// noopXABuilder 使用空行为实现 XABuilder 接口。
type noopXABuilder struct{}

func (b *noopXABuilder) Add(url string, payload any) integrationcontract.XABuilder {
	_, _ = url, payload
	return b
}

func (b *noopXABuilder) Submit(ctx context.Context) error {
	_ = ctx
	return ErrNoopDTM
}

// noopBarrierHandler implements BarrierHandler with no-op behavior.
// 注意：Call 方法返回 ErrNoopDTM，不会调用业务回调 fn。
// 这与其他 noop provider 的"静默安全"策略一致。
//
// noopBarrierHandler 使用空行为实现 BarrierHandler 接口。
type noopBarrierHandler struct{}

func (h *noopBarrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	_ = ctx
	_ = fn
	// 返回 ErrNoopDTM 而非调用 fn(nil)，避免业务代码 nil panic
	// 与其他 noop provider 的"静默安全"策略一致
	return ErrNoopDTM
}
