package bootstrap

import (
	"github.com/ngq/gorp/framework"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	"github.com/ngq/gorp/framework/provider/orm/ent"
)

func DefaultProviders() []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0, 24)
	providers = append(providers, FoundationProviders()...)
	providers = append(providers, ORMRuntimeProviders()...)
	providers = append(providers, DefaultCapabilityProviders()...)
	return providers
}

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

func registerCLIProviders(c runtimecontract.Container) error {
	return c.RegisterProvider(ent.NewProvider())
}

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
