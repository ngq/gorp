package bootstrap

import (
	"github.com/ngq/gorp/framework/contract"
	appProvider "github.com/ngq/gorp/framework/provider/app"
	authJWTProvider "github.com/ngq/gorp/framework/provider/auth/jwt"
	configProvider "github.com/ngq/gorp/framework/provider/config"
	cronProvider "github.com/ngq/gorp/framework/provider/cron"
	ginProvider "github.com/ngq/gorp/framework/provider/gin"
	hostProvider "github.com/ngq/gorp/framework/provider/host"
	logProvider "github.com/ngq/gorp/framework/provider/log"
	gormProvider "github.com/ngq/gorp/framework/provider/orm/gorm"
	inspectProvider "github.com/ngq/gorp/framework/provider/orm/inspect"
	runtimeORMProvider "github.com/ngq/gorp/framework/provider/orm/runtime"
	sqlxProvider "github.com/ngq/gorp/framework/provider/orm/sqlx"
	redisProvider "github.com/ngq/gorp/framework/provider/redis"
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
// - 当前只保留 redis + 业务 JWT 这类不会与项目侧 runtime/capability selector 冲突的能力；
// - serviceauth 改由项目侧或统一主链路选择器负责，避免重复注册。
func BusinessSimplificationProviders() []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0, 8)
	providers = append(providers, redisProvider.NewProvider())
	providers = append(providers, AuthProviders()...)
	return providers
}
