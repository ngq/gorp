// Application scenarios:
// - Build the framework's default bootstrap application for CLI and general startup paths.
// - Reuse one shared provider assembly strategy for non-HTTP bootstrap entrypoints.
// - Expose a small set of helper constructors for default application creation.
//
// 适用场景：
// - 为 CLI 和通用启动路径构建框架默认 bootstrap application。
// - 为非 HTTP 的 bootstrap 入口复用统一的 provider 装配策略。
// - 暴露一组默认 application 构建辅助入口。
package bootstrap

import (
	"github.com/ngq/gorp/framework"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/provider/orm/ent"
)

// DefaultProviders returns the default provider set used by bootstrap helpers.
//
// DefaultProviders 返回 bootstrap helper 使用的默认 provider 集合。
func DefaultProviders() []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0, 24)
	providers = append(providers, FoundationProviders()...)
	providers = append(providers, ORMRuntimeProviders()...)
	providers = append(providers, DefaultCapabilityProviders()...)
	return providers
}

// NewCLIApplication builds a default application with CLI-oriented providers.
//
// NewCLIApplication 构建带有 CLI 场景 provider 的默认 application。
func NewCLIApplication() (*framework.Application, runtimecontract.Container, error) {
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

// registerCLIProviders registers CLI-only providers on top of the default application.
//
// registerCLIProviders 在默认 application 之上注册 CLI 专用 provider。
func registerCLIProviders(c runtimecontract.Container) error {
	return c.RegisterProvider(ent.NewProvider())
}

// NewDefaultApplication builds the default application used by general bootstrap flows.
//
// NewDefaultApplication 构建通用 bootstrap 流程使用的默认 application。
func NewDefaultApplication() (*framework.Application, runtimecontract.Container, error) {
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

// DefaultBootstrapHelpers returns the helper names commonly surfaced to users.
//
// DefaultBootstrapHelpers 返回通常对外暴露的 bootstrap helper 名称列表。
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
