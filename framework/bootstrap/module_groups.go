// Application scenarios:
// - Group providers into stable bootstrap bundles by responsibility.
// - Give application assembly code predictable provider slices for foundation, ORM, auth, and business helpers.
// - Keep top-level bootstrap composition declarative instead of scattering provider lists.
//
// 适用场景：
// - 按职责把 provider 组织成稳定的 bootstrap 分组。
// - 为应用装配代码提供 foundation、ORM、auth 和业务简化能力的可预测 provider 切片。
// - 让顶层 bootstrap 组合保持声明式，而不是把 provider 列表散落各处。
package bootstrap

import (
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	appProvider "github.com/ngq/gorp/framework/provider/app"
	authJWTProvider "github.com/ngq/gorp/framework/provider/auth/jwt"
	cacheProvider "github.com/ngq/gorp/framework/provider/cache"
	configProvider "github.com/ngq/gorp/framework/provider/config"
	cronProvider "github.com/ngq/gorp/framework/provider/cron"
	ginProvider "github.com/ngq/gorp/framework/provider/gin"
	hostProvider "github.com/ngq/gorp/framework/provider/host"
	logProvider "github.com/ngq/gorp/framework/provider/log"
	gormProvider "github.com/ngq/gorp/framework/provider/orm/gorm"
	inspectProvider "github.com/ngq/gorp/framework/provider/orm/inspect"
	runtimeORMProvider "github.com/ngq/gorp/framework/provider/orm/runtime"
	sqlxProvider "github.com/ngq/gorp/framework/provider/orm/sqlx"
)

// FoundationProviders returns the baseline providers required by most applications.
//
// FoundationProviders 返回大多数应用都需要的基础 provider 集合。
func FoundationProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		appProvider.NewProvider(),
		configProvider.NewProvider(),
		logProvider.NewProvider(),
		ginProvider.NewProvider(),
		hostProvider.NewProvider(),
		cronProvider.NewProvider(),
	}
}

// CoreProviders is an alias for FoundationProviders.
//
// CoreProviders 是 FoundationProviders 的同义入口。
func CoreProviders() []runtimecontract.ServiceProvider {
	return FoundationProviders()
}

// ORMRuntimeProviders returns the providers that expose ORM and DB runtime capabilities.
//
// ORMRuntimeProviders 返回暴露 ORM 与数据库运行时能力的 provider 集合。
func ORMRuntimeProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		gormProvider.NewProvider(),
		sqlxProvider.NewProvider(),
		runtimeORMProvider.NewProvider(),
		inspectProvider.NewProvider(),
	}
}

// AuthProviders returns the providers used for business-side authentication.
//
// AuthProviders 返回业务侧认证使用的 provider 集合。
func AuthProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		authJWTProvider.NewProvider(),
	}
}

// ServiceAuthProviders returns the providers used for service-to-service auth.
//
// ServiceAuthProviders 返回服务间认证使用的 provider 集合。
func ServiceAuthProviders() []runtimecontract.ServiceProvider {
	return nil
}

// BusinessSimplificationProviders returns providers that simplify common business access patterns.
//
// BusinessSimplificationProviders 返回用于简化常见业务接入的 provider 集合。
func BusinessSimplificationProviders() []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0, 8)
	providers = append(providers, AuthProviders()...)
	providers = append(providers, cacheProvider.NewProvider())
	return providers
}

// DefaultCapabilityProviders returns the default non-foundation capability providers.
//
// DefaultCapabilityProviders 返回默认的非基础能力 provider 集合。
func DefaultCapabilityProviders() []runtimecontract.ServiceProvider {
	return BusinessSimplificationProviders()
}
