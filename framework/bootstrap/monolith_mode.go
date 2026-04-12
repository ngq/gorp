package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
	cbnoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	cfglocal "github.com/ngq/gorp/framework/provider/configsource/local"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	protonoopp "github.com/ngq/gorp/framework/provider/proto/noop"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
	validatenoop "github.com/ngq/gorp/framework/provider/validate/noop"
	retrynoop "github.com/ngq/gorp/framework/provider/retry/noop"
)

// MonolithFriendlyProviders 返回“单体友好模式”下推荐注册的微服务能力 provider 集合。
//
// 中文说明：
// - 这是 framework 级统一入口，用于让业务项目快速进入单体友好模式；
// - 目标不是关闭框架能力，而是为服务发现、RPC、服务认证、追踪、消息队列等组件提供零依赖/noop 实现；
// - 这样业务项目不必逐个理解 discovery/rpc/serviceauth/tracing 等组件该选择哪个 noop provider；
// - 配置源默认使用 local，以便单体项目直接从本地 config 启动。
func MonolithFriendlyProviders() []contract.ServiceProvider {
	return []contract.ServiceProvider{
		cfglocal.NewProvider(),
		discoverynoop.NewProvider(),
		rpcnoop.NewProvider(),
		serviceauthnoop.NewProvider(),
		tracingnoop.NewProvider(),
		mqnoop.NewProvider(),
		dlocknoop.NewProvider(),
		cbnoop.NewProvider(),
		selectornoop.NewProvider(),
		metadatanoop.NewProvider(),
		validatenoop.NewProvider(),
		retrynoop.NewProvider(),
		protonoopp.NewProvider(),
	}
}
