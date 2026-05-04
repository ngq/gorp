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
