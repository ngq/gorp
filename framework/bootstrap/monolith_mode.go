// Application scenarios:
// - Provide a safe default provider bundle for monolith-friendly deployments.
// - Prefer noop/local implementations for capabilities that are optional in single-process apps.
// - Let monolith assembly stay simple without losing contract completeness.
//
// 适用场景：
// - 为单体友好部署提供安全的默认 provider 组合。
// - 对单进程应用中可选的能力优先使用 noop/local 实现。
// - 让 monolith 装配保持简单，同时不丢失契约完整性。
package bootstrap

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	cbnoop "github.com/ngq/gorp/framework/provider/circuitbreaker/noop"
	cfglocal "github.com/ngq/gorp/framework/provider/configsource/local"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dlocknoop "github.com/ngq/gorp/framework/provider/dlock/noop"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	mqnoop "github.com/ngq/gorp/framework/provider/messagequeue/noop"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	protonoopp "github.com/ngq/gorp/framework/provider/proto/noop"
	retrynoop "github.com/ngq/gorp/framework/provider/retry/noop"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
	validatenoop "github.com/ngq/gorp/framework/provider/validate/noop"
)

// MonolithFriendlyProviders returns the provider bundle tailored for monolith deployments.
//
// MonolithFriendlyProviders 返回面向单体部署的 provider 组合。
func MonolithFriendlyProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		cfglocal.NewProvider(),
		discoverynoop.NewProvider(),
		rpcnoop.NewProvider(),
		serviceauthnoop.NewProvider(),
		tracingnoop.NewProvider(),
		mqnoop.NewProvider(),
		dlocknoop.NewProvider(),
		cbnoop.NewProvider(),
		dtmnoop.NewProvider(),
		selectornoop.NewProvider(),
		metadatanoop.NewProvider(),
		validatenoop.NewProvider(),
		retrynoop.NewProvider(),
		protonoopp.NewProvider(),
	}
}
