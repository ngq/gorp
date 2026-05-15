// Package ent provides Ent ORM integration for gorp framework.
// Ent requires project-level code generation, so the provider expects a custom factory.
// Unlike GORM/SQLX, Ent client must be provided by the application:
//
// Ent ORM 包，提供基于 Facebook Ent 的代码生成式 ORM 能力。
// Ent 需要项目级代码生成，因此 Provider 期望自定义工厂。
// 与 GORM/SQLX 不同，Ent 客户端必须由应用提供：
//   - Bind EntClientFactoryKey before registering this provider
//   - Or provide the generated client via custom factory
//
// 注册此 Provider 前需绑定 EntClientFactoryKey，或通过自定义工厂提供生成的客户端。
// Configuration via config.yaml:
//
// 通过 config.yaml 配置：
//
//	database:
//	  backend: ent
//	  driver: mysql
//	  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4"
//
// Eg:
//
//	// 应用层绑定工厂
//	c.Bind(datacontract.EntClientFactoryKey, &MyEntFactory{})
//	app.Register(ent.NewProvider())
package ent

import (
	"fmt"

	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

// Provider registers the Ent client contract.
//
// Provider 注册 Ent 客户端契约。
type Provider struct{}

// NewProvider creates a new Ent provider instance.
//
// NewProvider 创建新的 Ent Provider 实例。
func NewProvider() *Provider { return &Provider{} }

// Name returns the provider name "orm.ent".
//
// Name 返回 Provider 名称 "orm.ent"。
func (p *Provider) Name() string { return "orm.ent" }

// IsDefer returns false, Ent should be initialized immediately for DB connection.
//
// IsDefer 返回 false，Ent 应立即初始化以建立数据库连接。
func (p *Provider) IsDefer() bool { return false }

// Provides returns the Ent client contract key.
//
// Provides 返回 Ent 客户端契约键。
func (p *Provider) Provides() []string { return []string{datacontract.EntClientKey} }

// DependsOn returns the keys this provider depends on.
// Ent provider depends on Config for database configuration.
//
// DependsOn 返回该 provider 依赖的 key。
// Ent provider 依赖 Config 获取数据库配置。
func (p *Provider) DependsOn() []string { return []string{datacontract.ConfigKey} }

// Register binds the Ent client factory to the container.
// Core logic: Resolve factory from container, call CreateEntClient to get client.
// Note: Application must bind EntClientFactoryKey before registering this provider.
//
// Register 将 Ent 客户端工厂绑定到容器。
// 核心逻辑：从容器解析工厂，调用 CreateEntClient 获取客户端。
// 注意：应用必须在注册此 Provider 前绑定 EntClientFactoryKey。
func (p *Provider) Register(c runtimecontract.Container) error {
	c.Bind(datacontract.EntClientKey, func(c runtimecontract.Container) (any, error) {
		factoryAny, err := c.Make(datacontract.EntClientFactoryKey)
		if err != nil {
			cfgAny, cfgErr := c.Make(datacontract.ConfigKey)
			if cfgErr != nil {
				return nil, err
			}
			cfg := cfgAny.(datacontract.Config)
			var dbc datacontract.DBConfig
			_ = cfg.Unmarshal("database", &dbc)
			return nil, fmt.Errorf("database.backend=ent is selected, but no project-level ent factory is bound at %q (driver=%s)", datacontract.EntClientFactoryKey, dbc.Driver)
		}

		factory, ok := factoryAny.(datacontract.EntClientFactory)
		if !ok {
			return nil, fmt.Errorf("resolved %q does not implement contract.EntClientFactory", datacontract.EntClientFactoryKey)
		}
		return factory.CreateEntClient(c)
	}, true)
	return nil
}

// Boot is a no-op for Ent provider.
//
// Boot Ent Provider 无启动逻辑。
func (p *Provider) Boot(runtimecontract.Container) error { return nil }