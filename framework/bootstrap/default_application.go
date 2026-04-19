package bootstrap

import (
	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/contract"
)

// DefaultProviders 返回一组可直接用于业务项目起步的默认 provider。
//
// 中文说明：
// - 这是 framework 默认接入路径的唯一基础 provider 组；
// - 目标是让不使用模板的业务项目，也能用一组清晰的默认 provider 快速启动；
// - 当前提供的是：app/config/log/gin/host/cron/gorm/sqlx/orm.runtime/inspect/redis/auth.jwt；
// - 微服务能力不再通过 DefaultProviders() 全量竞争注册，而是在配置可用后由 capability selector 显式选中；
// - 若业务项目需要单体友好模式，应在此基础上继续注册 `MonolithFriendlyProviders()`，而不是改用别的并列入口。
func DefaultProviders() []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0, 24)
	providers = append(providers, FoundationProviders()...)
	providers = append(providers, ORMRuntimeProviders()...)
	providers = append(providers, BusinessSimplificationProviders()...)
	return providers
}

// NewDefaultApplication 创建一个已注册默认 provider 的 Application。
//
// 中文说明：
// - 这是给不使用模板的业务项目准备的更直接接入 helper；
// - 相比手工逐个 RegisterProvider，这里把默认启动骨架再收一层；
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
// - 后续如果 helper 按能力域继续收口，这里也应同步更新，而不是在文档里散落多份清单。
func DefaultBootstrapHelpers() []string {
	return []string{
		"container.MakeConfig",
		"container.MustMakeJWTService",
		"container.MustMakeAppService",
		"container.MakeDBRuntime",
		"container.MakeRedis",
		"container.MakeCache",
	}
}
