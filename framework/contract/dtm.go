package contract

import (
	"context"
)

const (
	// DTMKey 是 DTM 分布式事务客户端在容器中的绑定 key。
	//
	// 中文说明：
	// - 用于分布式事务编排；
	// - 支持 SAGA、TCC、XA 等模式；
	// - noop 实现空操作，单体项目零依赖；
	// - 需要独立部署 DTM Server（https://dtm.pub）。
	DTMKey = "framework.dtm"
)

// DTMClient DTM 分布式事务客户端接口。
//
// 中文说明：
// - 与 DTM Server 交互的客户端；
// - 支持 SAGA、TCC、XA 等分布式事务模式；
// - 框架只提供客户端 SDK，不实现事务逻辑。
type DTMClient interface {
	// SAGA 创建 SAGA 事务。
	//
	// 中文说明：
	// - SAGA 模式：将长事务拆分为多个本地事务；
	// - 每个步骤有对应的补偿操作；
	// - 适合最终一致性场景。
	SAGA(name string) SAGABuilder

	// TCC 创建 TCC 事务。
	//
	// 中文说明：
	// - TCC 模式：Try-Confirm-Cancel；
	// - Try：预留资源；
	// - Confirm：确认提交；
	// - Cancel：取消预留；
	// - 适合强一致性场景。
	TCC(name string) TCCBuilder

	// XA 创建 XA 事务。
	//
	// 中文说明：
	// - XA 模式：两阶段提交；
	// - 数据库需要支持 XA 协议；
	// - 适合强一致性场景。
	XA(name string) XABuilder

	// Barrier 创建 Barrier 事务（用于 TCC 自动补偿）。
	//
	// 中文说明：
	// - 自动处理空补偿、悬挂等问题；
	// - 简化 TCC 开发。
	Barrier(transType, gid string) BarrierHandler

	// Query 查询事务状态。
	Query(ctx context.Context, gid string) (*TransactionInfo, error)
}

// SAGABuilder SAGA 事务构建器。
//
// 中文说明：
// - 链式构建 SAGA 事务步骤；
// - 每个步骤包含正向操作和补偿操作。
type SAGABuilder interface {
	// Add 添加一个步骤。
	//
	// 中文说明：
	// - action: 正向操作 URL；
	// - compensate: 补偿操作 URL；
	// - payload: 请求数据。
	Add(action string, compensate string, payload any) SAGABuilder

	// AddBranch 添加分支（带事务分支选项）。
	AddBranch(action string, compensate string, payload any, opts BranchOptions) SAGABuilder

	// Submit 提交事务。
	Submit(ctx context.Context) error

	// Build 构建事务对象（不提交）。
	Build() (*SAGATransaction, error)
}

// TCCBuilder TCC 事务构建器。
type TCCBuilder interface {
	// Add 添加一个 TCC 分支。
	//
	// 中文说明：
	// - try: Try 操作 URL；
	// - confirm: Confirm 操作 URL；
	// - cancel: Cancel 操作 URL；
	// - payload: 请求数据。
	Add(try string, confirm string, cancel string, payload any) TCCBuilder

	// Submit 提交事务。
	Submit(ctx context.Context) error
}

// XABuilder XA 事务构建器。
type XABuilder interface {
	// Add 添加一个 XA 分支。
	//
	// 中文说明：
	// - url: 操作 URL；
	// - payload: 请求数据。
	Add(url string, payload any) XABuilder

	// Submit 提交事务。
	Submit(ctx context.Context) error
}

// BarrierHandler Barrier 事务处理器。
//
// 中文说明：
// - 自动处理幂等性、空补偿、悬挂；
// - 用于 TCC/MSG 模式。
type BarrierHandler interface {
	// Call 执行 Barrier 保护的操作。
	//
	// 中文说明：
	// - 自动处理并发问题；
	// - fn: 业务操作函数。
	Call(ctx context.Context, fn func(db any) error) error
}

// TransactionInfo 事务信息。
type TransactionInfo struct {
	// GID 全局事务 ID
	GID string

	// Status 事务状态
	Status string

	// TransactionType 事务类型：saga/tcc/xa/msg
	TransactionType string

	// CreateTime 创建时间
	CreateTime int64

	// UpdateTime 更新时间
	UpdateTime int64

	// Steps 步骤信息
	Steps []TransactionStep
}

// TransactionStep 事务步骤。
type TransactionStep struct {
	// BranchID 分支 ID
	BranchID string

	// Status 步骤状态
	Status string

	// Op 操作类型
	Op string

	// URL 操作 URL
	URL string
}

// SAGATransaction SAGA 事务对象。
type SAGATransaction struct {
	// GID 全局事务 ID
	GID string

	// Steps 步骤列表
	Steps []SAGAStep

	// Payloads 请求数据列表
	Payloads []any
}

// SAGAStep SAGA 步骤。
type SAGAStep struct {
	// Action 正向操作 URL
	Action string

	// Compensate 补偿操作 URL
	Compensate string

	// Payload 请求数据
	Payload any

	// RetryCount 分支级重试次数
	RetryCount int

	// RetryInterval 分支级重试间隔（秒）
	RetryInterval int

	// Timeout 分支级超时时间（秒）
	Timeout int
}

// BranchOptions 分支选项。
type BranchOptions struct {
	// RetryCount 重试次数
	RetryCount int

	// RetryInterval 重试间隔（秒）
	RetryInterval int

	// Timeout 超时时间（秒）
	Timeout int
}

// DTMConfig DTM 配置。
type DTMConfig struct {
	// Enabled 是否启用
	Enabled bool

	// DTM Server 地址
	Endpoint string

	// HTTP 超时时间（秒）
	Timeout int

	// 重试配置
	RetryCount    int
	RetryInterval int

	// 回调配置
	CallbackPort    int    // 回调端口
	CallbackAddress string // 回调地址（外网可访问）
}