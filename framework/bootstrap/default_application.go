package bootstrap

import (
	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/provider/orm/ent"
)

// DefaultProviders 返回一组可直接用于业务项目起步的默认 provider。
//
// 中文说明：
// - 这是 framework 默认接入路径的“default runtime skeleton + default capability”组合入口；
// - 目标是让不使用模板的业务项目，也能用一组清晰的 core runtime + default provider 快速启动；
// - 当前提供的是：CoreProviders + ORMRuntimeProviders + DefaultCapabilityProviders；
// - 微服务能力不再通过 DefaultProviders() 全量竞争注册，而是在配置可用后由 capability selector 显式选中；
// - 若业务项目需要单体友好模式，应在此基础上继续注册 `MonolithFriendlyProviders()`，而不是改用别的并列入口。
func DefaultProviders() []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0, 24)
	providers = append(providers, FoundationProviders()...)
	providers = append(providers, ORMRuntimeProviders()...)
	providers = append(providers, DefaultCapabilityProviders()...)
	return providers
}

// NewCLIApplication 创建一套供工具链内部复用的默认 Application + Container。
//
// 中文说明：
// - 这是给母仓 CLI / 工具链命令准备的装配入口；
// - 负责注册 default runtime skeleton、按配置选中微服务 provider，并补充 ORM runtime provider；
// - 业务项目默认启动入口仍应优先走自己生成出来的 `cmd/*/main.go`，而不是直接依赖这层 helper；
// - 这层 helper 明确属于 toolchain 场景，不是 framework 未来极简公共 root 的候选形态。
func NewCLIApplication() (*framework.Application, contract.Container, error) {
	app := framework.NewApplication()
	c := app.Container()
	if err := c.RegisterProviders(DefaultProviders()...); err != nil {
		return nil, nil, err
	}
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		return nil, nil, err
	}
	if err := registerCLIProviders(c); err != nil {
		return nil, nil, err
	}
	return app, c, nil
}

// registerCLIProviders 为 CLI / 工具链场景补齐额外 provider。
//
// 中文说明：
// - 这里只放母仓工具链确实需要的补充能力；
// - 目前仅补 ent provider，目的是让模板治理、生成校验等命令在母仓内具备完整 ORM runtime 解析能力；
// - 不把这层 helper 继续扩展成业务默认入口，避免 CLI 装配边界重新变宽。
func registerCLIProviders(c contract.Container) error {
	return c.RegisterProvider(ent.NewProvider())
}

// NewDefaultApplication 创建一个已注册默认 provider 的 Application。
//
// 中文说明：
// - 这是给不使用模板的业务项目准备的更直接接入 helper；
// - 相比手工逐个 RegisterProvider，这里把默认启动骨架再收一层；
// - 当前它仍明确属于 default runtime skeleton helper，而不是 framework 已冻结的极简公共 root；
// - 这里默认只注册 default runtime skeleton + default capability，再由 capability selector 按配置选择微服务主链路能力；
// - 单体友好模式下，可以在此基础上继续注册 MonolithFriendlyProviders。
func NewDefaultApplication() (*framework.Application, contract.Container, error) {
	app := framework.NewApplication()
	c := app.Container()
	if err := c.RegisterProviders(DefaultProviders()...); err != nil {
		return nil, nil, err
	}
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		return nil, nil, err
	}
	return app, c, nil
}

// DefaultBootstrapHelpers 暴露当前推荐使用的 container helper 集合说明。
//
// 中文说明：
// - 这里只返回字符串列表，主要用于文档与默认接入路径统一口径；
// - 这些 helper 属于“默认推荐的业务接入 helper”，而不是唯一合法入口；
// - 这里列出的应优先是 typed runtime 之外、默认业务路径仍然合理可见的轻入口；
// - 后续如果 helper 按能力域继续收口，这里也应同步更新，而不是在文档里散落多份清单。
func DefaultBootstrapHelpers() []string {
	return []string{
		"container.MakeConfig",
		"container.MakeDBRuntime",
		"container.MakeRedis",
		"container.MakeCache",
		"container.MakeValidator",
		"container.MakeRetry",
	}
}
