// Application scenarios:
// - Provide the full provider bundle used by microservice-oriented bootstrap paths.
// - Collect discovery, selector, RPC, tracing, metadata, service-auth, and DTM capabilities in one place.
// - Keep microservice mainline assembly explicit and reusable.
//
// 适用场景：
// - 为微服务导向的 bootstrap 路径提供完整 provider 组合。
// - 在一个位置集中 discovery、selector、RPC、tracing、metadata、service-auth 和 DTM 能力。
// - 让微服务主线装配保持显式且可复用。
package bootstrap

import (
	configsourceconsul "github.com/ngq/gorp/contrib/configsource/consul"
	configsourceetcd "github.com/ngq/gorp/contrib/configsource/etcd"
	dtmsdk "github.com/ngq/gorp/contrib/dtm/dtmsdk"
	discoveryconsul "github.com/ngq/gorp/contrib/registry/consul"
	discoveryetcd "github.com/ngq/gorp/contrib/registry/etcd"
	discoverynacos "github.com/ngq/gorp/contrib/registry/nacos"
	serviceauthmtls "github.com/ngq/gorp/contrib/serviceauth/mtls"
	serviceauthtoken "github.com/ngq/gorp/contrib/serviceauth/token"
	tracingotel "github.com/ngq/gorp/contrib/tracing/otel"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
)

// MicroserviceMainlineProviders returns the provider bundle used by the microservice mainline.
//
// MicroserviceMainlineProviders 返回微服务主线使用的 provider 组合。
func MicroserviceMainlineProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		configsourcelocal.NewProvider(),
		configsourceconsul.NewProvider(),
		configsourceetcd.NewProvider(),
		configsourcenoop.NewProvider(),
		discoveryconsul.NewProvider(),
		discoveryetcd.NewProvider(),
		discoverynacos.NewProvider(),
		discoverynoop.NewProvider(),
		selectornoop.NewProvider(),
		selectorrandom.NewProvider(),
		selectorwrr.NewProvider(),
		selectorp2c.NewProvider(),
		rpcnoop.NewProvider(),
		rpchttp.NewProvider(),
		rpcgrpc.NewProvider(),
		tracingnoop.NewProvider(),
		tracingotel.NewProvider(),
		metadatanoop.NewProvider(),
		metadatadefault.NewProvider(),
		serviceauthnoop.NewProvider(),
		serviceauthtoken.NewProvider(),
		serviceauthmtls.NewProvider(),
		dtmnoop.NewProvider(),
		dtmsdk.NewProvider(),
	}
}
