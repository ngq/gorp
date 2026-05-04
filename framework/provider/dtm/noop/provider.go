package noop

import (
	"context"
	"errors"

	integrationcontract "github.com/ngq/gorp/framework/contract/integration"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string { return "dtm.noop" }

func (p *Provider) IsDefer() bool { return true }

func (p *Provider) Provides() []string { return []string{integrationcontract.DTMKey} }

func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(integrationcontract.DTMKey, func(c runtimecontract.Container) (any, error) {
		return &noopDTMClient{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(runtimecontract.Container) error { return nil }

var ErrNoopDTM = errors.New("dtm: noop mode, distributed transaction not available in monolith")

type noopDTMClient struct{}

func (c *noopDTMClient) SAGA(name string) integrationcontract.SAGABuilder {
	_ = name
	return &noopSAGABuilder{}
}

func (c *noopDTMClient) TCC(name string) integrationcontract.TCCBuilder {
	_ = name
	return &noopTCCBuilder{}
}

func (c *noopDTMClient) XA(name string) integrationcontract.XABuilder {
	_ = name
	return &noopXABuilder{}
}

func (c *noopDTMClient) Barrier(transType, gid string) integrationcontract.BarrierHandler {
	_ = transType
	_ = gid
	return &noopBarrierHandler{}
}

func (c *noopDTMClient) Query(ctx context.Context, gid string) (*integrationcontract.TransactionInfo, error) {
	_ = ctx
	_ = gid
	return nil, ErrNoopDTM
}

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

type noopTCCBuilder struct{}

func (b *noopTCCBuilder) Add(try string, confirm string, cancel string, payload any) integrationcontract.TCCBuilder {
	_, _, _, _ = try, confirm, cancel, payload
	return b
}

func (b *noopTCCBuilder) Submit(ctx context.Context) error {
	_ = ctx
	return ErrNoopDTM
}

type noopXABuilder struct{}

func (b *noopXABuilder) Add(url string, payload any) integrationcontract.XABuilder {
	_, _ = url, payload
	return b
}

func (b *noopXABuilder) Submit(ctx context.Context) error {
	_ = ctx
	return ErrNoopDTM
}

type noopBarrierHandler struct{}

func (h *noopBarrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	_ = ctx
	return fn(nil)
}
