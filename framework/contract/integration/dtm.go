// Application scenarios:
// - Define the distributed transaction contract used by business orchestration flows.
// - Support SAGA, TCC, XA, and barrier-style transaction coordination behind one abstraction.
// - Keep DTM client and transaction configuration provider-neutral.
//
// 适用场景：
// - 定义业务编排流程使用的分布式事务契约。
// - 在统一抽象下支持 SAGA、TCC、XA 和 barrier 风格的事务协调。
// - 保持 DTM client 和事务配置与具体 provider 解耦。
package integration

import "context"

// DTMKey is the container key for the distributed transaction capability.
//
// DTMKey 是分布式事务能力的容器键。
const DTMKey = "framework.dtm"

// DTMClient defines the distributed transaction client contract.
//
// DTMClient 定义分布式事务客户端契约。
type DTMClient interface {
	SAGA(name string) SAGABuilder
	TCC(name string) TCCBuilder
	XA(name string) XABuilder
	Barrier(transType, gid string) BarrierHandler
	Query(ctx context.Context, gid string) (*TransactionInfo, error)
}

// SAGABuilder defines the SAGA transaction builder contract.
//
// SAGABuilder 定义 SAGA 事务构建器契约。
type SAGABuilder interface {
	Add(action string, compensate string, payload any) SAGABuilder
	AddBranch(action string, compensate string, payload any, opts BranchOptions) SAGABuilder
	Submit(ctx context.Context) error
	Build() (*SAGATransaction, error)
}

// TCCBuilder defines the TCC transaction builder contract.
//
// TCCBuilder 定义 TCC 事务构建器契约。
type TCCBuilder interface {
	Add(try string, confirm string, cancel string, payload any) TCCBuilder
	Submit(ctx context.Context) error
}

// XABuilder defines the XA transaction builder contract.
//
// XABuilder 定义 XA 事务构建器契约。
type XABuilder interface {
	Add(url string, payload any) XABuilder
	Submit(ctx context.Context) error
}

// BarrierHandler defines the transaction barrier execution contract.
//
// BarrierHandler 定义事务屏障执行契约。
type BarrierHandler interface {
	Call(ctx context.Context, fn func(db any) error) error
}

// DTMConfig describes distributed transaction runtime configuration.
//
// DTMConfig 描述分布式事务运行时配置。
type DTMConfig struct {
	Enabled         bool
	Endpoint        string
	Timeout         int
	RetryCount      int
	RetryInterval   int
	CallbackPort    int
	CallbackAddress string
}
