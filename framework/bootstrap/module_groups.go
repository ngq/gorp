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

func CoreProviders() []runtimecontract.ServiceProvider {
	return FoundationProviders()
}

func ORMRuntimeProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		gormProvider.NewProvider(),
		sqlxProvider.NewProvider(),
		runtimeORMProvider.NewProvider(),
		inspectProvider.NewProvider(),
	}
}

func AuthProviders() []runtimecontract.ServiceProvider {
	return []runtimecontract.ServiceProvider{
		authJWTProvider.NewProvider(),
	}
}

func ServiceAuthProviders() []runtimecontract.ServiceProvider {
	return nil
}

func BusinessSimplificationProviders() []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0, 8)
	providers = append(providers, AuthProviders()...)
	providers = append(providers, cacheProvider.NewProvider())
	return providers
}

func DefaultCapabilityProviders() []runtimecontract.ServiceProvider {
	return BusinessSimplificationProviders()
}
