// Application scenarios:
// - Centralize governance-mode detection used by bootstrap selection and startup helpers.
// - Keep monolith and microservice default-provider behavior explicit in one place.
// - Allow application options and config-driven assembly to share the same runtime-mode semantics.
//
// 适用场景：
// - 集中处理 bootstrap 选择与启动 helper 使用的治理模式判断。
// - 将 monolith 与 microservice 的默认 provider 语义显式收口到一个位置。
// - 让 application 选项与配置驱动装配共享同一套运行模式语义。
package bootstrap

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
)

// DetectGovernanceMode resolves the runtime governance mode from config.
//
// DetectGovernanceMode 从配置中解析运行时治理模式。
func DetectGovernanceMode(cfg datacontract.Config) resiliencecontract.GovernanceMode {
	if cfg == nil {
		return resiliencecontract.GovernanceModeMonolith
	}

	for _, key := range []string{
		"governance.mode",
		"app.mode",
		"runtime.mode",
		"service.mode",
	} {
		switch cfg.GetString(key) {
		case string(resiliencecontract.GovernanceModeMicroservice):
			return resiliencecontract.GovernanceModeMicroservice
		case string(resiliencecontract.GovernanceModeMonolith):
			return resiliencecontract.GovernanceModeMonolith
		}
	}

	return resiliencecontract.GovernanceModeMonolith
}

// NormalizeGovernanceMode returns a supported governance mode, defaulting to monolith.
//
// NormalizeGovernanceMode 返回受支持的治理模式；未命中时默认回退为 monolith。
func NormalizeGovernanceMode(mode resiliencecontract.GovernanceMode) resiliencecontract.GovernanceMode {
	switch mode {
	case resiliencecontract.GovernanceModeMicroservice:
		return resiliencecontract.GovernanceModeMicroservice
	default:
		return resiliencecontract.GovernanceModeMonolith
	}
}

// IsMicroserviceMode reports whether the given mode represents the microservice mainline.
//
// IsMicroserviceMode 返回当前模式是否代表微服务主线。
func IsMicroserviceMode(mode resiliencecontract.GovernanceMode) bool {
	return NormalizeGovernanceMode(mode) == resiliencecontract.GovernanceModeMicroservice
}
