package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
	configsourceconsul "github.com/ngq/gorp/framework/provider/configsource/consul"
	configsourcelocal "github.com/ngq/gorp/framework/provider/configsource/local"
	configsourcenoop "github.com/ngq/gorp/framework/provider/configsource/noop"
	configsourceetcd "github.com/ngq/gorp/framework/provider/configsource/etcd"
	discoveryconsul "github.com/ngq/gorp/framework/provider/discovery/consul"
	discoveryetcd "github.com/ngq/gorp/framework/provider/discovery/etcd"
	discoverynacos "github.com/ngq/gorp/framework/provider/discovery/nacos"
	discoverynoop "github.com/ngq/gorp/framework/provider/discovery/noop"
	dtmnoop "github.com/ngq/gorp/framework/provider/dtm/noop"
	dtmsdk "github.com/ngq/gorp/framework/provider/dtm/dtmsdk"
	metadatadefault "github.com/ngq/gorp/framework/provider/metadata"
	metadatanoop "github.com/ngq/gorp/framework/provider/metadata/noop"
	rpchttp "github.com/ngq/gorp/framework/provider/rpc/http"
	rpcgrpc "github.com/ngq/gorp/framework/provider/rpc/grpc"
	rpcnoop "github.com/ngq/gorp/framework/provider/rpc/noop"
	selectornoop "github.com/ngq/gorp/framework/provider/selector/noop"
	selectorp2c "github.com/ngq/gorp/framework/provider/selector/p2c"
	selectorrandom "github.com/ngq/gorp/framework/provider/selector/random"
	selectorwrr "github.com/ngq/gorp/framework/provider/selector/wrr"
	serviceauthmtls "github.com/ngq/gorp/framework/provider/serviceauth/mtls"
	serviceauthnoop "github.com/ngq/gorp/framework/provider/serviceauth/noop"
	serviceauthtoken "github.com/ngq/gorp/framework/provider/serviceauth/token"
	tracingnoop "github.com/ngq/gorp/framework/provider/tracing/noop"
	tracingotel "github.com/ngq/gorp/framework/provider/tracing/otel"
)

// MicroserviceMainlineProviders 返回当前主链路推荐 provider 集合。
//
// 中文说明：
// - 这是“允许参与主链路选择的 provider 白名单”；
// - 当前保留此函数，主要给过渡阶段与内部验证使用；
// - 默认业务主线后续应逐步改为通过 capability selector 显式选中单一 provider，而不是继续依赖注册顺序决定行为。
func MicroserviceMainlineProviders() []contract.ServiceProvider {
	return []contract.ServiceProvider{
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
