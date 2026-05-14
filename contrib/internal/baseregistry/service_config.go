// Package baseregistry provides shared service config reading utilities.
// These helpers standardize the reading of common discovery.service.*
// and selector.algorithm configuration across all registry providers.
//
// baseregistry 包提供共享的服务配置读取工具。
// 这些辅助工具标准化所有 registry provider 中 discovery.service.*
// 和 selector.algorithm 的配置读取。
package baseregistry

import (
	datacontract "github.com/ngq/gorp/framework/contract/data"
	configprovider "github.com/ngq/gorp/framework/provider/config"
)

// ServiceConfig holds the common service registration fields shared by all registry providers.
//
// ServiceConfig 保存所有 registry provider 共享的通用服务注册字段。
type ServiceConfig struct {
	ServiceName string
	ServiceAddr string
	ServicePort int
	ServiceMeta map[string]string
	LoadBalance string
}

// ReadServiceConfig reads the common discovery.service.* and selector.algorithm configuration.
//
// ReadServiceConfig 读取通用的 discovery.service.* 和 selector.algorithm 配置。
func ReadServiceConfig(cfg datacontract.Config) ServiceConfig {
	sc := ServiceConfig{}
	sc.ServiceName = configprovider.GetStringAny(cfg,
		"discovery.service.name",
		"discovery.service_name",
	)
	sc.ServiceAddr = configprovider.GetStringAny(cfg,
		"discovery.service.addr",
		"discovery.service.address",
		"discovery.service_addr",
	)
	sc.ServicePort = configprovider.GetIntAny(cfg,
		"discovery.service.port",
		"discovery.service_port",
	)
	sc.LoadBalance = configprovider.GetStringAny(cfg,
		"selector.algorithm",
		"discovery.load_balance",
	)
	return sc
}
