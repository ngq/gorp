package runtime_provider

import (
	"fmt"

	apphttp "github.com/ngq/gorp/app/http"
	appservice "github.com/ngq/gorp/app/service"
	"github.com/ngq/gorp/framework/contract"

	"google.golang.org/grpc"
)

// HTTPRuntimeConfigurator 收口母仓 legacy HTTP runtime 命令组的装配动作。
type HTTPRuntimeConfigurator struct{}

// CronRuntimeConfigurator 收口母仓 legacy cron runtime 命令组的装配动作。
type CronRuntimeConfigurator struct{}

// GRPCRuntimeBuilder 收口母仓 legacy gRPC runtime 命令组的 server 构造动作。
type GRPCRuntimeBuilder struct{}

// Provider 注册母仓自己的 runtime provider。
//
// 中文说明：
// - 这是母仓自带的 runtime provider，不代表 generated starter 的默认公开启动方式；
// - 它主要服务于母仓保留的 legacy/runtime 命令组与兼容验证链路；
// - 项目模板如需覆盖这套装配，会通过自己的 runtime provider 注入共享 CLI bootstrap。
type Provider struct{}

func NewProvider() *Provider { return &Provider{} }

func (p *Provider) Name() string  { return "app.runtime" }
func (p *Provider) IsDefer() bool { return false }
func (p *Provider) Provides() []string {
	return []string{contract.HTTPRuntimeConfiguratorKey, contract.CronRuntimeConfiguratorKey, contract.GRPCRuntimeBuilderKey}
}

func (p *Provider) Register(c contract.Container) error {
	c.Bind(contract.HTTPRuntimeConfiguratorKey, func(contract.Container) (any, error) {
		return HTTPRuntimeConfigurator{}, nil
	}, true)
	c.Bind(contract.CronRuntimeConfiguratorKey, func(contract.Container) (any, error) {
		return CronRuntimeConfigurator{}, nil
	}, true)
	c.Bind(contract.GRPCRuntimeBuilderKey, func(contract.Container) (any, error) {
		return GRPCRuntimeBuilder{}, nil
	}, true)
	return nil
}

func (p *Provider) Boot(contract.Container) error { return nil }

func (HTTPRuntimeConfigurator) ConfigureHTTPRuntime(c contract.Container) error {
	if err := appservice.AutoMigrate(c); err != nil {
		return err
	}
	if err := apphttp.RegisterRoutes(c); err != nil {
		return err
	}
	return nil
}

func (CronRuntimeConfigurator) ConfigureCronRuntime(c contract.Container) (int, error) {
	return 0, nil
}

func (GRPCRuntimeBuilder) BuildGRPCServer() *grpc.Server {
	panic(fmt.Errorf("grpc runtime is not configured in the current mother repo; use external examples to demonstrate gRPC capabilities"))
}
