package noop

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 noop DTM 分布式事务实现。
//
// 中文说明：
// - 单体项目默认使用此 provider；
// - 不引入任何外部依赖；
// - 所有分布式事务操作为空实现；
// - 单体项目使用本地事务，不需要分布式事务。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "dtm.noop" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.DTMKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DTMKey, func(c contract.Container) (any, error) {
		return &noopDTMClient{}, nil
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// ErrNoopDTM 表示 noop 模式不支持分布式事务。
var ErrNoopDTM = errors.New("dtm: noop mode, distributed transaction not available in monolith")

// noopDTMClient 是 DTMClient 的空实现。
type noopDTMClient struct{}

// SAGA 创建 SAGA 事务（返回空构建器）。
func (c *noopDTMClient) SAGA(name string) contract.SAGABuilder {
	return &noopSAGABuilder{}
}

// TCC 创建 TCC 事务（返回空构建器）。
func (c *noopDTMClient) TCC(name string) contract.TCCBuilder {
	return &noopTCCBuilder{}
}

// XA 创建 XA 事务（返回空构建器）。
func (c *noopDTMClient) XA(name string) contract.XABuilder {
	return &noopXABuilder{}
}

// Barrier 创建 Barrier 事务（返回空处理器）。
func (c *noopDTMClient) Barrier(transType, gid string) contract.BarrierHandler {
	return &noopBarrierHandler{}
}

// Query 查询事务状态（返回错误）。
func (c *noopDTMClient) Query(ctx context.Context, gid string) (*contract.TransactionInfo, error) {
	return nil, ErrNoopDTM
}

// noopSAGABuilder 是 SAGABuilder 的空实现。
type noopSAGABuilder struct{}

func (b *noopSAGABuilder) Add(action string, compensate string, payload any) contract.SAGABuilder {
	return b
}

func (b *noopSAGABuilder) AddBranch(action string, compensate string, payload any, opts contract.BranchOptions) contract.SAGABuilder {
	return b
}

func (b *noopSAGABuilder) Submit(ctx context.Context) error {
	return ErrNoopDTM
}

func (b *noopSAGABuilder) Build() (*contract.SAGATransaction, error) {
	return nil, ErrNoopDTM
}

// noopTCCBuilder 是 TCCBuilder 的空实现。
type noopTCCBuilder struct{}

func (b *noopTCCBuilder) Add(try string, confirm string, cancel string, payload any) contract.TCCBuilder {
	return b
}

func (b *noopTCCBuilder) Submit(ctx context.Context) error {
	return ErrNoopDTM
}

// noopXABuilder 是 XABuilder 的空实现。
type noopXABuilder struct{}

func (b *noopXABuilder) Add(url string, payload any) contract.XABuilder {
	return b
}

func (b *noopXABuilder) Submit(ctx context.Context) error {
	return ErrNoopDTM
}

// noopBarrierHandler 是 BarrierHandler 的空实现。
type noopBarrierHandler struct{}

func (h *noopBarrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	// 直接执行函数（无 Barrier 保护）
	return fn(nil)
}