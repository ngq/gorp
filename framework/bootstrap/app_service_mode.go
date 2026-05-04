package bootstrap

import runtimecontract "github.com/ngq/gorp/framework/contract/runtime"

func RegisterAppServices(c runtimecontract.Container, bindings map[string]runtimecontract.Factory) {
	for key, factory := range bindings {
		if factory == nil {
			continue
		}
		c.Bind(key, factory, true)
	}
}
