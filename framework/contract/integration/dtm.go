package integration

import "context"

const DTMKey = "framework.dtm"

type DTMClient interface {
	SAGA(name string) SAGABuilder
	TCC(name string) TCCBuilder
	XA(name string) XABuilder
	Barrier(transType, gid string) BarrierHandler
	Query(ctx context.Context, gid string) (*TransactionInfo, error)
}

type SAGABuilder interface {
	Add(action string, compensate string, payload any) SAGABuilder
	AddBranch(action string, compensate string, payload any, opts BranchOptions) SAGABuilder
	Submit(ctx context.Context) error
	Build() (*SAGATransaction, error)
}

type TCCBuilder interface {
	Add(try string, confirm string, cancel string, payload any) TCCBuilder
	Submit(ctx context.Context) error
}

type XABuilder interface {
	Add(url string, payload any) XABuilder
	Submit(ctx context.Context) error
}

type BarrierHandler interface {
	Call(ctx context.Context, fn func(db any) error) error
}

type DTMConfig struct {
	Enabled         bool
	Endpoint        string
	Timeout         int
	RetryCount      int
	RetryInterval   int
	CallbackPort    int
	CallbackAddress string
}
