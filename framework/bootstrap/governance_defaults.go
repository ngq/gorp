// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file centralizes mode-aware default provider backends for governance bootstrap.
// Provides stable source of truth for monolith, microservice, and gin-first defaults.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件集中维护治理 bootstrap 路径里按模式生效的默认 provider backend。
// 为 monolith、microservice、gin-first 三种模式提供统一默认值真源。
package bootstrap

import resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"

// GovernanceProviderDefaults describes the mode-aware default provider backends.
//
// GovernanceProviderDefaults 描述按治理模式生效的默认 provider backend。
type GovernanceProviderDefaults struct {
	ConfigSource    string
	Discovery       string
	Selector        string
	RPC             string
	Tracing         string
	Metadata        string
	ServiceAuth     string
	CircuitBreaker  string
	LoadShedder     string
	Retry           string
	DTM             string
	MessageQueue    string
	DistributedLock string
}

// DefaultGovernanceProviderDefaults returns the default provider backend bundle for one governance mode.
//
// DefaultGovernanceProviderDefaults 返回某个治理模式对应的默认 provider backend 组合。
func DefaultGovernanceProviderDefaults(mode resiliencecontract.GovernanceMode) GovernanceProviderDefaults {
	mode = NormalizeGovernanceMode(mode)
	features := resiliencecontract.DefaultGovernanceFeatureSet(mode)

	defaults := GovernanceProviderDefaults{
		ConfigSource:    "local",
		Discovery:       "noop",
		Selector:        "noop",
		RPC:             "noop",
		Tracing:         "noop",
		Metadata:        "noop",
		ServiceAuth:     "noop",
		CircuitBreaker:  "noop",
		LoadShedder:     "noop",
		Retry:           "noop",
		DTM:             "noop",
		MessageQueue:    "noop",
		DistributedLock: "noop",
	}

	if features.Discovery {
		defaults.Discovery = "etcd"
	}
	if features.Selector {
		defaults.Selector = "p2c_ewma"
	}
	if features.Tracing {
		defaults.Tracing = "otel"
	}
	if features.MetadataPropagation {
		defaults.Metadata = "default"
	}
	if features.ServiceAuth {
		defaults.ServiceAuth = "token"
	}
	if features.CircuitBreaker {
		defaults.CircuitBreaker = "sentinel"
	}
	if features.LoadShedding {
		defaults.LoadShedder = "semaphore"
	}

	return defaults
}
