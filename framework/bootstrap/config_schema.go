// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file defines critical config schemas and fail-fast validation at startup.
// Uses go-playground/validator to catch missing or invalid config values early,
// instead of letting them silently become viper zero values that cause confusing runtime errors.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件定义关键配置的结构体 schema 与启动时 fail-fast 校验。
// 使用 go-playground/validator 在启动阶段就捕获缺失或无效的配置值，
// 避免它们被 viper 静默转为零值后引发难以排查的运行时错误。
package bootstrap

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	datacontract "github.com/ngq/gorp/framework/contract/data"
)

// ---------------------------------------------------------------------------
// 关键配置结构体定义
// 每个结构体对应 YAML 配置文件中的一个关键节（section），
// 带 validate tag 用于 go-playground/validator 校验。
// ---------------------------------------------------------------------------

// AppConfigSchema 对应 YAML 配置中的 app 节。
// 校验 app.address（HTTP 监听地址），这是服务启动的必要条件。
//
// AppConfigSchema maps the "app" section in YAML config.
// Validates app.address (HTTP listen address), which is required for service startup.
type AppConfigSchema struct {
	// Address 是 HTTP 服务监听地址，必填，例如 ":8080" 或 "0.0.0.0:8080"。
	//
	// Address is the HTTP listen address, required, e.g. ":8080" or "0.0.0.0:8080".
	Address string `validate:"required" mapstructure:"address"`
}

// ServerHTTPConfigSchema 对应 YAML 配置中的 server.http 节。
// 当用户显式配置 server.http 时，addr 必须提供。
//
// ServerHTTPConfigSchema maps the "server.http" section in YAML config.
// When server.http is explicitly configured, addr must be provided.
type ServerHTTPConfigSchema struct {
	// Addr 是 HTTP 服务监听地址，必填（当 server.http 节存在时）。
	//
	// Addr is the HTTP listen address, required when server.http section is present.
	Addr string `validate:"required" mapstructure:"addr"`
}

// LogConfigSchema 对应 YAML 配置中的 log 节。
// 日志级别和格式是服务可观测性的基础，缺失会导致日志行为不可预期。
//
// LogConfigSchema maps the "log" section in YAML config.
// Log level and format are foundational for observability;
// missing values lead to unpredictable logging behavior.
type LogConfigSchema struct {
	// Level 是日志级别，必填，可选值：debug / info / warn / error。
	//
	// Level is the log level, required. Valid values: debug / info / warn / error.
	Level string `validate:"required,oneof=debug info warn error" mapstructure:"level"`

	// Format 是日志输出格式，必填，可选值：console / json。
	//
	// Format is the log output format, required. Valid values: console / json.
	Format string `validate:"required,oneof=console json" mapstructure:"format"`
}

// DatabaseConfigSchema 对应 YAML 配置中的 database 节。
// 数据库配置是条件必填：仅当 database 节存在且 driver/dsn 非空时才校验。
// 若用户未配置 database，校验自动跳过（不强制使用数据库）。
//
// DatabaseConfigSchema maps the "database" section in YAML config.
// Database config is conditionally required: validated only when the database
// section exists and driver/dsn are non-empty.
// If the user does not configure database, validation is skipped (DB is optional).
type DatabaseConfigSchema struct {
	// Driver 是数据库驱动，可选值：sqlite / mysql / postgres / pgx。
	// 当 DSN 非空时，Driver 也必须提供。
	//
	// Driver is the database driver. Valid values: sqlite / mysql / postgres / pgx.
	// When DSN is non-empty, Driver is also required.
	Driver string `validate:"required_with=DSN,oneof=sqlite mysql postgres pgx" mapstructure:"driver"`

	// DSN 是数据库连接字符串，当 Driver 非空时必须提供。
	//
	// DSN is the database connection string, required when Driver is non-empty.
	DSN string `validate:"required_with=Driver" mapstructure:"dsn"`
}

// TracingConfigSchema 对应 YAML 配置中的 tracing 节。
// 仅当 tracing.enabled=true 时校验，缺失或关闭时跳过。
//
// TracingConfigSchema maps the "tracing" section in YAML config.
// Validated only when tracing.enabled=true; skipped when absent or disabled.
type TracingConfigSchema struct {
	// Backend 是追踪后端，启用时必填，可选值：otel / zipkin / noop。
	Backend string `validate:"required,oneof=otel zipkin noop" mapstructure:"backend"`

	// ServiceName 是追踪服务名，启用时必填。
	ServiceName string `validate:"required" mapstructure:"service_name"`
}

// RedisConfigSchema 对应 YAML 配置中的 redis 节。
// 仅当 redis 段存在且 addr 非空时校验。
//
// RedisConfigSchema maps the "redis" section in YAML config.
// Validated only when the section exists and addr is non-empty.
type RedisConfigSchema struct {
	// Addr 是 Redis 连接地址，使用 Redis 时必填。
	Addr string `validate:"required" mapstructure:"addr"`
}

// ServiceAuthConfigSchema 对应 YAML 配置中的 service_auth 节。
// 仅当 service_auth.enabled=true 时校验。
//
// ServiceAuthConfigSchema maps the "service_auth" section in YAML config.
// Validated only when service_auth.enabled=true.
type ServiceAuthConfigSchema struct {
	// Backend 是服务间认证后端，启用时必填，可选值：token / mtls / noop。
	Backend string `validate:"required,oneof=token mtls noop" mapstructure:"backend"`
}

// ---------------------------------------------------------------------------
// 校验逻辑
// ---------------------------------------------------------------------------

// configValidator 是包级共享的 validator 实例，避免每次校验都重新创建。
// 使用 validator.New() 初始化，不注册结构体缓存以保持简单。
//
// configValidator is the package-level shared validator instance.
// Initialized with validator.New(), no struct cache registration for simplicity.
var configValidator *validator.Validate

func init() {
	configValidator = validator.New()
	// 使用结构体字段名（而非 tag 名）作为错误信息中的字段标识，
	// 这样用户看到的错误更接近 YAML 配置键名。
	//
	// Use struct field names (not tag names) in error messages
	// so the field names are closer to the YAML config keys users write.
	configValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("mapstructure"), ",", 2)[0]
		if name == "" {
			name = fld.Name
		}
		return name
	})
}

// ValidateCriticalConfig 对关键配置节执行 fail-fast 校验。
// 校验策略：
//   - 始终校验：app（HTTP 监听地址）、log（级别和格式）
//   - 条件校验：server.http（当节存在时）、database（当节存在且非空时）
//   - 跳过：tracing / redis 等可选节（缺失时不校验）
//
// 返回值：
//   - nil：所有关键配置校验通过
//   - error：包含所有校验失败项的聚合错误，每条错误带节名前缀
//
// ValidateCriticalConfig performs fail-fast validation on critical config sections.
// Strategy:
//   - Always validate: app (HTTP listen addr), log (level & format)
//   - Conditional: server.http (when section exists), database (when section exists & non-empty)
//   - Skip: optional sections like tracing, redis (not validated when absent)
//
// Returns:
//   - nil: all critical config sections passed validation
//   - error: aggregated error with section name prefix on each field error
func ValidateCriticalConfig(cfg datacontract.Config) error {
	var errs []string

	// 1. 校验 app 节（始终校验，HTTP 监听地址是服务启动的必要条件）
	// 1. Validate "app" section (always; HTTP listen address is required for service startup)
	var appCfg AppConfigSchema
	if err := cfg.Unmarshal("app", &appCfg); err != nil {
		errs = append(errs, fmt.Sprintf("config: failed to unmarshal app section: %v", err))
	} else if fieldErrs := configValidator.Struct(appCfg); fieldErrs != nil {
		errs = append(errs, formatSectionErrors("app", fieldErrs)...)
	}

	// 2. 校验 server.http 节（条件校验：仅在配置中存在时才校验）
	// 2. Validate "server.http" section (conditional: only when present in config)
	if cfg.Get("server.http") != nil {
		var httpCfg ServerHTTPConfigSchema
		if err := cfg.Unmarshal("server.http", &httpCfg); err != nil {
			errs = append(errs, fmt.Sprintf("config: failed to unmarshal server.http section: %v", err))
		} else if fieldErrs := configValidator.Struct(httpCfg); fieldErrs != nil {
			errs = append(errs, formatSectionErrors("server.http", fieldErrs)...)
		}
	}

	// 3. 校验 log 节（始终校验，日志级别和格式是可观测性基础）
	// 3. Validate "log" section (always; level and format are foundational for observability)
	var logCfg LogConfigSchema
	if err := cfg.Unmarshal("log", &logCfg); err != nil {
		errs = append(errs, fmt.Sprintf("config: failed to unmarshal log section: %v", err))
	} else if fieldErrs := configValidator.Struct(logCfg); fieldErrs != nil {
		errs = append(errs, formatSectionErrors("log", fieldErrs)...)
	}

	// 4. 校验 database 节（条件校验：仅在节存在且 driver 或 dsn 非空时才校验）
	// 若 database 节完全缺失或 driver/dsn 都为空字符串，跳过校验。
	// 4. Validate "database" section (conditional: only when section exists and driver/dsn non-empty)
	// If the database section is entirely absent or both driver/dsn are empty, skip validation.
	if cfg.Get("database") != nil {
		var dbCfg DatabaseConfigSchema
		if err := cfg.Unmarshal("database", &dbCfg); err != nil {
			errs = append(errs, fmt.Sprintf("config: failed to unmarshal database section: %v", err))
		} else if dbCfg.Driver != "" || dbCfg.DSN != "" {
			// 仅当 driver 或 dsn 至少有一个非空时才校验
			// Only validate when at least one of driver/dsn is non-empty
			if fieldErrs := configValidator.Struct(dbCfg); fieldErrs != nil {
				errs = append(errs, formatSectionErrors("database", fieldErrs)...)
			}
		}
	}

	// 5. 校验 tracing 节（条件校验：仅当 tracing.enabled=true 时才校验）
	// 5. Validate "tracing" section (conditional: only when tracing.enabled=true)
	if cfg.GetBool("tracing.enabled") {
		var traceCfg TracingConfigSchema
		if err := cfg.Unmarshal("tracing", &traceCfg); err != nil {
			errs = append(errs, fmt.Sprintf("config: failed to unmarshal tracing section: %v", err))
		} else if fieldErrs := configValidator.Struct(traceCfg); fieldErrs != nil {
			errs = append(errs, formatSectionErrors("tracing", fieldErrs)...)
		}
	}

	// 6. 校验 redis 节（条件校验：仅当 redis 段存在且 addr 非空时才校验）
	// 6. Validate "redis" section (conditional: only when section exists and addr non-empty)
	if cfg.Get("redis") != nil {
		var redisCfg RedisConfigSchema
		if err := cfg.Unmarshal("redis", &redisCfg); err != nil {
			errs = append(errs, fmt.Sprintf("config: failed to unmarshal redis section: %v", err))
		} else if redisCfg.Addr != "" {
			if fieldErrs := configValidator.Struct(redisCfg); fieldErrs != nil {
				errs = append(errs, formatSectionErrors("redis", fieldErrs)...)
			}
		}
	}

	// 7. 校验 service_auth 节（条件校验：仅当 service_auth.enabled=true 时才校验）
	// 7. Validate "service_auth" section (conditional: only when service_auth.enabled=true)
	if cfg.GetBool("service_auth.enabled") {
		var authCfg ServiceAuthConfigSchema
		if err := cfg.Unmarshal("service_auth", &authCfg); err != nil {
			errs = append(errs, fmt.Sprintf("config: failed to unmarshal service_auth section: %v", err))
		} else if fieldErrs := configValidator.Struct(authCfg); fieldErrs != nil {
			errs = append(errs, formatSectionErrors("service_auth", fieldErrs)...)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

// formatSectionErrors 将 validator.ValidationErrors 转换为带节名前缀的人性化错误消息。
// 输出格式为 "config: <section>.<field> <rule>"，例如：
//   - "config: app.address is required"
//   - "config: log.level must be one of [debug info warn error]"
//
// formatSectionErrors converts validator.ValidationErrors into human-friendly messages
// with section name prefix. Output format: "config: <section>.<field> <rule>", e.g.:
//   - "config: app.address is required"
//   - "config: log.level must be one of [debug info warn error]"
func formatSectionErrors(section string, err error) []string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// 非 ValidationErrors 类型，直接返回原始错误文本
		// Not a ValidationErrors type, return raw error text
		return []string{fmt.Sprintf("config: %s: %v", section, err)}
	}

	result := make([]string, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fieldName := fe.Field() // 通过 RegisterTagNameFunc 已映射为 mapstructure 名
		msg := formatFieldError(section, fieldName, fe)
		result = append(result, msg)
	}
	return result
}

// formatFieldError 将单条字段校验错误格式化为可读消息。
// 核心规则：
//   - required / required_with：提示字段缺失
//   - oneof：提示可选值列表
//   - 其他规则：显示原始 tag
//
// formatFieldError formats a single field validation error into a readable message.
// Core rules:
//   - required / required_with: indicate field is missing
//   - oneof: show the list of valid values
//   - other tags: show the raw tag
func formatFieldError(section, field string, fe validator.FieldError) string {
	key := fmt.Sprintf("config: %s.%s", section, field)
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", key)
	case "required_with":
		return fmt.Sprintf("%s is required (when %s is set)", key, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", key, fe.Param())
	default:
		return fmt.Sprintf("%s failed validation '%s'", key, fe.Tag())
	}
}
