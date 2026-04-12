package dtmsdk

import (
	"context"
	"errors"

	"github.com/ngq/gorp/framework/contract"
)

// Provider 提供 DTM 分布式事务客户端适配器骨架。
//
// 中文说明：
// - DTM 是成熟的分布式事务管理器（https://dtm.pub）；
// - 框架只提供契约和适配器骨架，不自研事务逻辑；
// - 使用此 provider 需要手动引入 DTM SDK：
//   go get github.com/dtm-labs/client
// - 需要独立部署 DTM Server。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string     { return "dtm.sdk" }
func (p *Provider) IsDefer() bool    { return true }
func (p *Provider) Provides() []string {
	return []string{contract.DTMKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.DTMKey, func(c contract.Container) (any, error) {
		cfg, err := getDTMConfig(c)
		if err != nil {
			return nil, err
		}
		return NewDTMClient(cfg)
	}, true)

	return nil
}

func (p *Provider) Boot(c contract.Container) error {
	return nil
}

// getDTMConfig 从容器获取 DTM 配置。
func getDTMConfig(c contract.Container) (*contract.DTMConfig, error) {
	cfgAny, err := c.Make(contract.ConfigKey)
	if err != nil {
		return nil, err
	}

	cfg, ok := cfgAny.(contract.Config)
	if !ok {
		return nil, errors.New("dtm: invalid config service")
	}

	dtmCfg := &contract.DTMConfig{
		Enabled:         true,
		Endpoint:        "http://localhost:36789",
		Timeout:         10,
		RetryCount:      3,
		RetryInterval:   5,
		CallbackPort:    8080,
		CallbackAddress: "localhost",
	}

	// DTM Server 地址
	if endpoint := cfg.GetString("dtm.endpoint"); endpoint != "" {
		dtmCfg.Endpoint = endpoint
	}

	// 是否启用
	if enabled := cfg.GetBool("dtm.enabled"); !enabled {
		dtmCfg.Enabled = false
	}

	// 超时配置
	if timeout := cfg.GetInt("dtm.timeout"); timeout > 0 {
		dtmCfg.Timeout = timeout
	}

	// 回调配置
	if port := cfg.GetInt("dtm.callback_port"); port > 0 {
		dtmCfg.CallbackPort = port
	}
	if addr := cfg.GetString("dtm.callback_address"); addr != "" {
		dtmCfg.CallbackAddress = addr
	}

	return dtmCfg, nil
}

// ErrDTMSDKNotImported 表示 DTM SDK 未引入。
var ErrDTMSDKNotImported = errors.New(`dtm: SDK not imported, please run:
  go get github.com/dtm-labs/client

For more details, see https://dtm.pub`)

// DTMClient 是 DTM 客户端适配器骨架。
//
// 中文说明：
// - 这是一个骨架实现，展示如何集成 DTM SDK；
// - 实际使用时需要引入 github.com/dtm-labs/client；
// - 参考 https://dtm.pub 获取完整使用指南。
type DTMClient struct {
	cfg *contract.DTMConfig
}

// NewDTMClient 创建 DTM 客户端。
func NewDTMClient(cfg *contract.DTMConfig) (*DTMClient, error) {
	return &DTMClient{
		cfg: cfg,
	}, nil
}

// SAGA 创建 SAGA 事务。
//
// 使用示例（需先引入 DTM SDK）：
//
//	import dtmclient "github.com/dtm-labs/client"
//
//	func example() {
//	    gid := dtmclient.MustGenGid("http://localhost:36789")
//	    saga := dtmclient.NewSaga("http://localhost:36789", gid)
//	    saga.Add("/api/order/create", "/api/order/compensate", orderReq)
//	    saga.Add("/api/inventory/deduct", "/api/inventory/restore", inventoryReq)
//	    err := saga.Submit()
//	}
func (c *DTMClient) SAGA(name string) contract.SAGABuilder {
	return &sagaBuilder{client: c, name: name}
}

// TCC 创建 TCC 事务。
//
// 使用示例（需先引入 DTM SDK）：
//
//	import dtmclient "github.com/dtm-labs/client"
//
//	func example() {
//	    gid := dtmclient.MustGenGid("http://localhost:36789")
//	    tcc := dtmclient.NewTCC("http://localhost:36789", gid)
//	    err := tcc.CallBranch(req, "/api/try", "/api/confirm", "/api/cancel")
//	}
func (c *DTMClient) TCC(name string) contract.TCCBuilder {
	return &tccBuilder{client: c, name: name}
}

// XA 创建 XA 事务。
func (c *DTMClient) XA(name string) contract.XABuilder {
	return &xaBuilder{client: c, name: name}
}

// Barrier 创建 Barrier 事务。
func (c *DTMClient) Barrier(transType, gid string) contract.BarrierHandler {
	return &barrierHandler{client: c, transType: transType, gid: gid}
}

// Query 查询事务状态。
func (c *DTMClient) Query(ctx context.Context, gid string) (*contract.TransactionInfo, error) {
	// TODO: 使用 DTM HTTP API 查询事务状态
	// GET {endpoint}/api/dtmsvr/query?gid={gid}
	return &contract.TransactionInfo{
		GID:            gid,
		Status:         "unknown",
		TransactionType: "unknown",
	}, nil
}

// sagaBuilder 是 SAGA 事务构建器骨架。
type sagaBuilder struct {
	client *DTMClient
	name   string
	steps  []sagaStep
}

type sagaStep struct {
	action     string
	compensate string
	payload    any
}

func (b *sagaBuilder) Add(action string, compensate string, payload any) contract.SAGABuilder {
	b.steps = append(b.steps, sagaStep{
		action:     action,
		compensate: compensate,
		payload:    payload,
	})
	return b
}

func (b *sagaBuilder) AddBranch(action string, compensate string, payload any, opts contract.BranchOptions) contract.SAGABuilder {
	return b.Add(action, compensate, payload)
}

func (b *sagaBuilder) Submit(ctx context.Context) error {
	// TODO: 引入 DTM SDK 后实现
	// saga := dtmclient.NewSaga(b.client.cfg.Endpoint, gid)
	// for _, step := range b.steps {
	//     saga.Add(step.action, step.compensate, step.payload)
	// }
	// return saga.Submit()
	return ErrDTMSDKNotImported
}

func (b *sagaBuilder) Build() (*contract.SAGATransaction, error) {
	steps := make([]contract.SAGAStep, len(b.steps))
	for i, step := range b.steps {
		steps[i] = contract.SAGAStep{
			Action:     step.action,
			Compensate: step.compensate,
			Payload:    step.payload,
		}
	}
	return &contract.SAGATransaction{Steps: steps}, nil
}

// tccBuilder 是 TCC 事务构建器骨架。
type tccBuilder struct {
	client *DTMClient
	name   string
	steps  []tccStep
}

type tccStep struct {
	try     string
	confirm string
	cancel  string
	payload any
}

func (b *tccBuilder) Add(try string, confirm string, cancel string, payload any) contract.TCCBuilder {
	b.steps = append(b.steps, tccStep{
		try:     try,
		confirm: confirm,
		cancel:  cancel,
		payload: payload,
	})
	return b
}

func (b *tccBuilder) Submit(ctx context.Context) error {
	// TODO: 引入 DTM SDK 后实现
	return ErrDTMSDKNotImported
}

// xaBuilder 是 XA 事务构建器骨架。
type xaBuilder struct {
	client *DTMClient
	name   string
	steps  []xaStep
}

type xaStep struct {
	url     string
	payload any
}

func (b *xaBuilder) Add(url string, payload any) contract.XABuilder {
	b.steps = append(b.steps, xaStep{url: url, payload: payload})
	return b
}

func (b *xaBuilder) Submit(ctx context.Context) error {
	// TODO: 引入 DTM SDK 后实现
	return ErrDTMSDKNotImported
}

// barrierHandler 是 Barrier 事务处理器骨架。
type barrierHandler struct {
	client    *DTMClient
	transType string
	gid       string
}

func (h *barrierHandler) Call(ctx context.Context, fn func(db any) error) error {
	// TODO: 引入 DTM SDK 后实现
	// barrier := workflow.NewBarrier(h.transType, h.gid, "", 0)
	// return barrier.Call(ctx, fn)
	return fn(nil)
}