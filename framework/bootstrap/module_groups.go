package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
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

// FoundationProviders 返回默认启动骨架里的 Foundation/App 组。
//
// 中文说明：
// - 这组负责应用壳、配置、日志、HTTP 宿主、进程宿主与任务调度基础能力；
// - 默认业务项目启动时通常都需要这组能力。
func FoundationProviders() []contract.ServiceProvider {
	return []contract.ServiceProvider{
		appProvider.NewProvider(),
		configProvider.NewProvider(),
		logProvider.NewProvider(),
		ginProvider.NewProvider(),
		hostProvider.NewProvider(),
		cronProvider.NewProvider(),
	}
}

// ORMRuntimeProviders 返回默认启动骨架里的 ORM/Runtime 组。
//
// 中文说明：
// - 统一包含 gorm/sqlx/runtime/inspect 四个 provider；
// - `orm.runtime` 负责 ORM capability key 绑定，避免 cmd 层重复绑定。
func ORMRuntimeProviders() []contract.ServiceProvider {
	return []contract.ServiceProvider{
		gormProvider.NewProvider(),
		sqlxProvider.NewProvider(),
		runtimeORMProvider.NewProvider(),
		inspectProvider.NewProvider(),
	}
}

// AuthProviders 返回默认启动骨架里的业务认证能力组。
//
// 中文说明：
// - 当前聚焦业务 JWT（AuthJWTKey）；
// - 与 ServiceAuthProviders 显式分组，避免身份认证语义混淆。
func AuthProviders() []contract.ServiceProvider {
	return []contract.ServiceProvider{
		authJWTProvider.NewProvider(),
	}
}

// ServiceAuthProviders 返回默认启动骨架里的服务间认证能力组。
//
// 中文说明：
// - 当前主链路正在收口阶段，serviceauth 不再固定塞进默认 provider 组；
// - 避免与模板项目/project-owned runtime 或 capability selector 发生重复注册；
// - 后续由项目侧 runtime 或统一 capability selector 决定使用 noop / token / mtls。
func ServiceAuthProviders() []contract.ServiceProvider {
	return nil
}

// BusinessSimplificationProviders 返回默认启动骨架里的业务减负能力组。
//
// 中文说明：
// - 这组面向业务开发默认可复用能力；
// - 当前聚焦不会改变 HTTP/ORM 主路径语义、但能明显降低业务起步成本的能力；
// - 目前包含：业务 JWT + 统一 cache；Redis 原语仍由项目侧按需接入。
func BusinessSimplificationProviders() []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0, 8)
	providers = append(providers, AuthProviders()...)
	providers = append(providers, cacheProvider.NewProvider())
	return providers
}
