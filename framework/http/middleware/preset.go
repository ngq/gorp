// Application scenarios:
// - Provide stable default middleware bundles for public, internal, and admin HTTP APIs.
// - Give business code a simple assembly entry without manually sorting middleware order.
// - Keep recommended middleware ordering and defaults centralized in one place.
//
// 适用场景：
// - 为对外、内网和管理 HTTP API 提供稳定的默认中间件组合。
// - 让业务侧无需手工排序中间件，也能直接完成装配。
// - 将推荐顺序与默认值集中维护在一个主线入口中。
package middleware

import (
	"runtime"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RecommendedMiddlewareOptions controls the recommended API middleware bundle.
//
// RecommendedMiddlewareOptions 用于控制推荐 API 中间件组合的装配方式。
type RecommendedMiddlewareOptions struct {
	EnableMetrics          bool
	EnableCompression      bool
	Timeout                time.Duration
	BodyLimitBytes         int64
	CORS                   *CORSOptions
	DisableSecurityHeaders bool
	SecurityHeaders        *SecurityHeadersOptions
	DisableLocale          bool
	Locale                 *LocaleOptions
}

// InternalMiddlewareOptions controls the internal API middleware bundle.
//
// InternalMiddlewareOptions 用于控制内网 API 中间件组合。
type InternalMiddlewareOptions struct {
	API RecommendedMiddlewareOptions
}

// AdminMiddlewareOptions controls the admin API middleware bundle.
//
// AdminMiddlewareOptions 用于控制管理 API 中间件组合。
type AdminMiddlewareOptions struct {
	API                  RecommendedMiddlewareOptions
	Allowlist            []string
	EnableAudit          bool
	DisableAudit         bool
	AuditOptions         AuditOptions
	RequireAuthorization bool
	DisableAuthorization bool
	RequireAnyRoles      []string
	RequireAllRoles      []string
}

// DefaultHTTPServiceGovernanceOptions controls the default HTTP service governance preset.
//
// DefaultHTTPServiceGovernanceOptions 用于控制默认 HTTP 服务治理预设。
type DefaultHTTPServiceGovernanceOptions struct {
	API                    RecommendedMiddlewareOptions
	RateLimiter            resiliencecontract.RateLimiter
	RateLimitResource      string
	CircuitBreaker         resiliencecontract.CircuitBreaker
	CircuitBreakerResource string
	MaxConcurrent          int
}

// DefaultRecommendedMiddlewareOptions returns the default recommended API options.
//
// DefaultRecommendedMiddlewareOptions 返回推荐对外 API 的默认配置。
func DefaultRecommendedMiddlewareOptions() RecommendedMiddlewareOptions {
	return RecommendedMiddlewareOptions{
		EnableMetrics:     true,
		EnableCompression: false,
		Timeout:           15 * time.Second,
		BodyLimitBytes:    2 << 20,
		SecurityHeaders: func() *SecurityHeadersOptions {
			opts := DefaultSecurityHeadersOptions()
			return &opts
		}(),
		Locale: func() *LocaleOptions {
			opts := DefaultLocaleOptions()
			return &opts
		}(),
	}
}

// DefaultInternalMiddlewareOptions returns the default internal API options.
//
// DefaultInternalMiddlewareOptions 返回内网 API 的默认配置。
func DefaultInternalMiddlewareOptions() InternalMiddlewareOptions {
	opts := DefaultRecommendedMiddlewareOptions()
	opts.DisableLocale = true
	opts.DisableSecurityHeaders = true
	opts.EnableCompression = false
	opts.CORS = nil
	return InternalMiddlewareOptions{API: opts}
}

// DefaultAdminMiddlewareOptions returns the default admin API options.
//
// DefaultAdminMiddlewareOptions 返回管理 API 的默认配置。
func DefaultAdminMiddlewareOptions() AdminMiddlewareOptions {
	internal := DefaultInternalMiddlewareOptions()
	return AdminMiddlewareOptions{
		API:                  internal.API,
		EnableAudit:          true,
		RequireAuthorization: true,
	}
}

// DefaultHTTPServiceGovernanceDefaults returns the default HTTP service governance options.
//
// DefaultHTTPServiceGovernanceDefaults 返回默认 HTTP 服务治理选项。
func DefaultHTTPServiceGovernanceDefaults() DefaultHTTPServiceGovernanceOptions {
	return DefaultHTTPServiceGovernanceOptions{
		API:           DefaultRecommendedMiddlewareOptions(),
		MaxConcurrent: runtime.GOMAXPROCS(0) * 100, // 默认并发限制：CPU核数 * 100
	}
}

// DefaultMiddleware returns the stable default middleware baseline.
//
// DefaultMiddleware 返回稳定的默认中间件基线。
func DefaultMiddleware(base observabilitycontract.Logger) transportcontract.Middleware {
	return Chain(DefaultMiddlewareSet(base)...)
}

// DefaultMiddlewareSet returns the stable default middleware baseline as an ordered slice.
//
// DefaultMiddlewareSet 以有序切片形式返回稳定的默认中间件基线。
//
// Included middleware:
// - RequestIdentity
// - LoggingMiddleware
// - RecoveryMiddleware
//
// 包含的中间件：
// - RequestIdentity
// - LoggingMiddleware
// - RecoveryMiddleware
func DefaultMiddlewareSet(base observabilitycontract.Logger) []transportcontract.Middleware {
	return []transportcontract.Middleware{
		RequestIdentity(),
		LoggingMiddleware(base),
		RecoveryMiddleware(),
	}
}

// RecommendedAPIMiddleware returns the recommended public API middleware preset.
//
// RecommendedAPIMiddleware 返回推荐的对外 API 中间件预设。
func RecommendedAPIMiddleware(base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) transportcontract.Middleware {
	return Chain(RecommendedAPIMiddlewareSet(base, opts)...)
}

// RecommendedAPIMiddlewareSet returns the recommended public API middleware preset as an ordered slice.
//
// RecommendedAPIMiddlewareSet 以有序切片形式返回推荐的对外 API 中间件预设。
//
// Included middleware:
// - DefaultMiddlewareSet
// - CORS, when enabled
// - SecurityHeaders, when enabled
// - Timeout, when configured
// - BodyLimit, when configured
// - Locale, when enabled
// - MetricsMiddleware, when enabled
// - Compression, when enabled
//
// 包含的中间件：
// - DefaultMiddlewareSet
// - 启用时追加 CORS
// - 启用时追加 SecurityHeaders
// - 配置后追加 Timeout
// - 配置后追加 BodyLimit
// - 启用时追加 Locale
// - 启用时追加 MetricsMiddleware
// - 启用时追加 Compression
func RecommendedAPIMiddlewareSet(base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) []transportcontract.Middleware {
	opts = normalizeRecommendedMiddlewareOptions(opts)

	middleware := make([]transportcontract.Middleware, 0, 10)
	middleware = append(middleware, DefaultMiddlewareSet(base)...)

	if opts.CORS != nil {
		middleware = append(middleware, CORS(*opts.CORS))
	}
	if !opts.DisableSecurityHeaders && opts.SecurityHeaders != nil {
		middleware = append(middleware, SecurityHeaders(*opts.SecurityHeaders))
	}
	if opts.Timeout > 0 {
		middleware = append(middleware, Timeout(opts.Timeout))
	}
	if opts.BodyLimitBytes > 0 {
		middleware = append(middleware, BodyLimit(opts.BodyLimitBytes))
	}
	if !opts.DisableLocale && opts.Locale != nil {
		middleware = append(middleware, Locale(*opts.Locale))
	}
	if opts.EnableMetrics {
		middleware = append(middleware, MetricsMiddleware())
	}
	if opts.EnableCompression {
		middleware = append(middleware, Compression())
	}
	return middleware
}

// InternalAPIMiddleware returns the recommended internal API middleware preset.
//
// InternalAPIMiddleware 返回推荐的内网 API 中间件预设。
func InternalAPIMiddleware(base observabilitycontract.Logger, opts InternalMiddlewareOptions) transportcontract.Middleware {
	return Chain(InternalAPIMiddlewareSet(base, opts)...)
}

// InternalAPIMiddlewareSet returns the recommended internal API middleware preset as an ordered slice.
//
// InternalAPIMiddlewareSet 以有序切片形式返回推荐的内网 API 中间件预设。
//
// Included middleware:
// - RecommendedAPIMiddlewareSet after internal defaults are applied
//
// 包含的中间件：
// - 应用内网默认配置后的 RecommendedAPIMiddlewareSet
func InternalAPIMiddlewareSet(base observabilitycontract.Logger, opts InternalMiddlewareOptions) []transportcontract.Middleware {
	opts = normalizeInternalMiddlewareOptions(opts)
	return RecommendedAPIMiddlewareSet(base, opts.API)
}

// AdminAPIMiddleware returns the recommended admin API middleware preset.
//
// AdminAPIMiddleware 返回推荐的管理 API 中间件预设。
func AdminAPIMiddleware(base observabilitycontract.Logger, opts AdminMiddlewareOptions) transportcontract.Middleware {
	return Chain(AdminAPIMiddlewareSet(base, opts)...)
}

// AdminAPIMiddlewareSet returns the recommended admin API middleware preset as an ordered slice.
//
// AdminAPIMiddlewareSet 以有序切片形式返回推荐的管理 API 中间件预设。
//
// Included middleware:
// - InternalAPIMiddlewareSet
// - IPAllowlist, when configured
// - RequireAuthorization, when enabled
// - RequireAnyRole, when configured
// - RequireAllRoles, when configured
// - AuditMiddleware, when enabled
//
// 包含的中间件：
// - InternalAPIMiddlewareSet
// - 配置后追加 IPAllowlist
// - 启用时追加 RequireAuthorization
// - 配置后追加 RequireAnyRole
// - 配置后追加 RequireAllRoles
// - 启用时追加 AuditMiddleware
func AdminAPIMiddlewareSet(base observabilitycontract.Logger, opts AdminMiddlewareOptions) []transportcontract.Middleware {
	opts = normalizeAdminMiddlewareOptions(opts)

	middleware := make([]transportcontract.Middleware, 0, 12)
	middleware = append(middleware, InternalAPIMiddlewareSet(base, InternalMiddlewareOptions{API: opts.API})...)

	if len(opts.Allowlist) > 0 {
		middleware = append(middleware, IPAllowlist(opts.Allowlist...))
	}
	if !opts.DisableAuthorization && opts.RequireAuthorization {
		middleware = append(middleware, RequireAuthorization())
	}
	if len(opts.RequireAnyRoles) > 0 {
		middleware = append(middleware, RequireAnyRole(opts.RequireAnyRoles...))
	}
	if len(opts.RequireAllRoles) > 0 {
		middleware = append(middleware, RequireAllRoles(opts.RequireAllRoles...))
	}
	if !opts.DisableAudit && opts.EnableAudit {
		middleware = append(middleware, AuditMiddleware(base, opts.AuditOptions))
	}
	return middleware
}

// UseDefaultMiddleware applies the default middleware baseline to the router.
//
// UseDefaultMiddleware 将默认中间件基线装配到路由器。
func UseDefaultMiddleware(router transportcontract.Router, base observabilitycontract.Logger) {
	if router == nil {
		return
	}
	router.Use(DefaultMiddlewareSet(base)...)
}

// UseRecommendedAPIMiddleware applies the recommended public API middleware preset to the router.
//
// UseRecommendedAPIMiddleware 将推荐的对外 API 中间件预设装配到路由器。
func UseRecommendedAPIMiddleware(router transportcontract.Router, base observabilitycontract.Logger, opts RecommendedMiddlewareOptions) {
	if router == nil {
		return
	}
	router.Use(RecommendedAPIMiddlewareSet(base, opts)...)
}

// UseInternalAPIMiddleware applies the recommended internal API middleware preset to the router.
//
// UseInternalAPIMiddleware 将推荐的内网 API 中间件预设装配到路由器。
func UseInternalAPIMiddleware(router transportcontract.Router, base observabilitycontract.Logger, opts InternalMiddlewareOptions) {
	if router == nil {
		return
	}
	router.Use(InternalAPIMiddlewareSet(base, opts)...)
}

// UseAdminAPIMiddleware applies the recommended admin API middleware preset to the router.
//
// UseAdminAPIMiddleware 将推荐的管理 API 中间件预设装配到路由器。
func UseAdminAPIMiddleware(router transportcontract.Router, base observabilitycontract.Logger, opts AdminMiddlewareOptions) {
	if router == nil {
		return
	}
	router.Use(AdminAPIMiddlewareSet(base, opts)...)
}

// DefaultHTTPServiceGovernanceOrder returns the stable logical order of the HTTP server governance middleware chain.
//
// The order here defines the formal execution contract: middleware registered earlier wraps all later middleware,
// so it runs first on the inbound path and last on the outbound path.
// Any change to this list is a breaking change to the governance chain contract.
//
// DefaultHTTPServiceGovernanceOrder 返回 HTTP 服务端治理中间件链的稳定逻辑顺序。
//
// 此顺序定义了正式的执行契约：先注册的中间件包裹后注册的所有中间件，
// 因此在入站路径上最先执行、在出站路径上最后执行。
// 对本列表的任何变更都是治理链契约的破坏性变更。
func DefaultHTTPServiceGovernanceOrder() []string {
	return []string{
		"request_identity",
		"logging",
		"recovery",
		"cors",
		"security_headers",
		"timeout",
		"load_shedding",
		"rate_limit",
		"circuit_breaker",
		"body_limit",
		"locale",
		"metrics",
		"compression",
	}
}

// DefaultHTTPServiceGovernancePreset returns the default HTTP service governance middleware preset.
//
// DefaultHTTPServiceGovernancePreset 返回默认 HTTP 服务治理中间件预设。
func DefaultHTTPServiceGovernancePreset(base observabilitycontract.Logger, opts DefaultHTTPServiceGovernanceOptions) transportcontract.Middleware {
	return Chain(DefaultHTTPServiceGovernanceSet(base, opts)...)
}

// DefaultHTTPServiceGovernanceSet returns the default HTTP service governance preset as an ordered slice.
//
// DefaultHTTPServiceGovernanceSet 以有序切片形式返回默认 HTTP 服务治理预设。
//
// Included middleware:
// - DefaultMiddlewareSet
// - CORS, when configured
// - SecurityHeaders, when enabled
// - Timeout, when configured
// - LoadShedding, when MaxConcurrent > 0
// - RateLimit, when a limiter is provided
// - CircuitBreaker, when a breaker is provided
// - BodyLimit, when configured
// - Locale, when enabled
// - MetricsMiddleware, when enabled
// - Compression, when enabled
//
// 包含的中间件：
// - DefaultMiddlewareSet
// - 配置后追加 CORS
// - 启用时追加 SecurityHeaders
// - 配置后追加 Timeout
// - MaxConcurrent 大于 0 时追加 LoadShedding
// - 提供 limiter 时追加 RateLimit
// - 提供 breaker 时追加 CircuitBreaker
// - 配置后追加 BodyLimit
// - 启用时追加 Locale
// - 启用时追加 MetricsMiddleware
// - 启用时追加 Compression
func DefaultHTTPServiceGovernanceSet(base observabilitycontract.Logger, opts DefaultHTTPServiceGovernanceOptions) []transportcontract.Middleware {
	opts = normalizeDefaultHTTPServiceGovernanceOptions(opts)

	middleware := make([]transportcontract.Middleware, 0, 12)
	middleware = append(middleware, DefaultMiddlewareSet(base)...)

	if opts.API.CORS != nil {
		middleware = append(middleware, CORS(*opts.API.CORS))
	}
	if !opts.API.DisableSecurityHeaders && opts.API.SecurityHeaders != nil {
		middleware = append(middleware, SecurityHeaders(*opts.API.SecurityHeaders))
	}
	if opts.API.Timeout > 0 {
		middleware = append(middleware, Timeout(opts.API.Timeout))
	}
	if opts.MaxConcurrent > 0 {
		middleware = append(middleware, LoadShedding(opts.MaxConcurrent))
	}
	if opts.RateLimiter != nil {
		middleware = append(middleware, RateLimit(opts.RateLimiter, opts.RateLimitResource))
	}
	if opts.CircuitBreaker != nil {
		middleware = append(middleware, CircuitBreaker(opts.CircuitBreaker, opts.CircuitBreakerResource))
	}
	if opts.API.BodyLimitBytes > 0 {
		middleware = append(middleware, BodyLimit(opts.API.BodyLimitBytes))
	}
	if !opts.API.DisableLocale && opts.API.Locale != nil {
		middleware = append(middleware, Locale(*opts.API.Locale))
	}
	if opts.API.EnableMetrics {
		middleware = append(middleware, MetricsMiddleware())
	}
	if opts.API.EnableCompression {
		middleware = append(middleware, Compression())
	}

	return middleware
}

// UseDefaultHTTPServiceGovernance applies the default HTTP service governance preset to the router.
//
// UseDefaultHTTPServiceGovernance 将默认 HTTP 服务治理预设装配到路由器。
func UseDefaultHTTPServiceGovernance(router transportcontract.Router, base observabilitycontract.Logger, opts DefaultHTTPServiceGovernanceOptions) {
	if router == nil {
		return
	}
	router.Use(DefaultHTTPServiceGovernanceSet(base, opts)...)
}

// normalizeRecommendedMiddlewareOptions fills missing recommended middleware options with defaults.
//
// normalizeRecommendedMiddlewareOptions 用默认值补齐推荐中间件选项。
func normalizeRecommendedMiddlewareOptions(opts RecommendedMiddlewareOptions) RecommendedMiddlewareOptions {
	defaults := DefaultRecommendedMiddlewareOptions()
	if opts == (RecommendedMiddlewareOptions{}) {
		return defaults
	}
	if opts.Timeout == 0 {
		opts.Timeout = defaults.Timeout
	}
	if opts.BodyLimitBytes == 0 {
		opts.BodyLimitBytes = defaults.BodyLimitBytes
	}
	if !opts.DisableSecurityHeaders && opts.SecurityHeaders == nil {
		opts.SecurityHeaders = defaults.SecurityHeaders
	}
	if !opts.DisableLocale && opts.Locale == nil {
		opts.Locale = defaults.Locale
	}
	return opts
}

// normalizeDefaultHTTPServiceGovernanceOptions fills missing HTTP governance preset options with defaults.
//
// normalizeDefaultHTTPServiceGovernanceOptions 用默认值补齐 HTTP 治理预设选项。
func normalizeDefaultHTTPServiceGovernanceOptions(opts DefaultHTTPServiceGovernanceOptions) DefaultHTTPServiceGovernanceOptions {
	defaults := DefaultHTTPServiceGovernanceDefaults()
	if opts == (DefaultHTTPServiceGovernanceOptions{}) {
		return defaults
	}
	opts.API = normalizeRecommendedMiddlewareOptions(opts.API)
	if opts.MaxConcurrent == 0 {
		opts.MaxConcurrent = defaults.MaxConcurrent
	}
	return opts
}

// normalizeInternalMiddlewareOptions fills missing internal middleware options with defaults.
//
// normalizeInternalMiddlewareOptions 用默认值补齐内网中间件选项。
func normalizeInternalMiddlewareOptions(opts InternalMiddlewareOptions) InternalMiddlewareOptions {
	defaults := DefaultInternalMiddlewareOptions()
	if opts == (InternalMiddlewareOptions{}) {
		return defaults
	}
	opts.API = normalizeRecommendedMiddlewareOptions(opts.API)
	if !opts.API.DisableLocale && defaults.API.DisableLocale {
		opts.API.DisableLocale = true
		opts.API.Locale = nil
	}
	if !opts.API.DisableSecurityHeaders && defaults.API.DisableSecurityHeaders {
		opts.API.DisableSecurityHeaders = true
		opts.API.SecurityHeaders = nil
	}
	return opts
}

// normalizeAdminMiddlewareOptions fills missing admin middleware options with defaults.
//
// normalizeAdminMiddlewareOptions 用默认值补齐管理中间件选项。
func normalizeAdminMiddlewareOptions(opts AdminMiddlewareOptions) AdminMiddlewareOptions {
	defaults := DefaultAdminMiddlewareOptions()
	if isZeroAdminMiddlewareOptions(opts) {
		return defaults
	}
	opts.API = normalizeInternalMiddlewareOptions(InternalMiddlewareOptions{API: opts.API}).API
	if !opts.DisableAudit && !opts.EnableAudit {
		opts.EnableAudit = defaults.EnableAudit
	}
	if !opts.DisableAuthorization && !opts.RequireAuthorization {
		opts.RequireAuthorization = defaults.RequireAuthorization
	}
	return opts
}

// isZeroAdminMiddlewareOptions reports whether admin options are still in their zero-value state.
//
// isZeroAdminMiddlewareOptions 判断管理中间件选项是否仍处于零值状态。
func isZeroAdminMiddlewareOptions(opts AdminMiddlewareOptions) bool {
	return opts.API == (RecommendedMiddlewareOptions{}) &&
		len(opts.Allowlist) == 0 &&
		!opts.EnableAudit &&
		!opts.DisableAudit &&
		isZeroAuditOptions(opts.AuditOptions) &&
		!opts.RequireAuthorization &&
		!opts.DisableAuthorization &&
		len(opts.RequireAnyRoles) == 0 &&
		len(opts.RequireAllRoles) == 0
}

// isZeroAuditOptions reports whether audit options are still in their zero-value state.
//
// isZeroAuditOptions 判断审计选项是否仍处于零值状态。
func isZeroAuditOptions(opts AuditOptions) bool {
	return opts.Event == "" &&
		opts.Action == nil &&
		opts.Resource == nil &&
		opts.Skip == nil
}

// Chain combines multiple middlewares into a single middleware.
//
// Chain 将多个中间件组合成一个中间件。
func Chain(middleware ...transportcontract.Middleware) transportcontract.Middleware {
	return func(next transportcontract.Handler) transportcontract.Handler {
		for i := len(middleware) - 1; i >= 0; i-- {
			if middleware[i] == nil {
				continue
			}
			next = middleware[i](next)
		}
		return next
	}
}
