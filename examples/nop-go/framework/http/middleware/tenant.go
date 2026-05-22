// Application scenarios:
// - Resolve tenant identity from route params, headers, or query parameters for multi-tenant APIs.
// - Propagate tenant information through request context for downstream business code and middleware.
// - Enforce required tenant presence at the HTTP mainline boundary.
//
// 适用场景：
// - 为多租户 API 从路由参数、请求头或查询参数中解析租户标识。
// - 通过请求上下文把租户信息传递给下游业务代码和中间件。
// - 在 HTTP 主线边界强制要求租户信息存在。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	supportcontract "github.com/ngq/gorp/framework/contract/support"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

const defaultTenantHeader = "X-Tenant-ID"

// TenantOptions controls request-level tenant resolution behavior.
//
// TenantOptions 控制请求级租户解析行为。
type TenantOptions struct {
	ParamKeys   []string
	HeaderKeys  []string
	QueryKeys   []string
	Default     string
	Required    bool
	WriteHeader bool
	HeaderName  string
}

// DefaultTenantOptions returns the default tenant resolution behavior.
//
// DefaultTenantOptions 返回默认的租户解析配置。
func DefaultTenantOptions() TenantOptions {
	return TenantOptions{
		ParamKeys:   []string{"tenant", "tenant_id"},
		HeaderKeys:  []string{defaultTenantHeader, "X-Tenant"},
		QueryKeys:   []string{"tenant", "tenant_id"},
		WriteHeader: false,
		HeaderName:  defaultTenantHeader,
	}
}

// Tenant resolves tenant identity and stores it in request context.
//
// Tenant 解析租户标识并将其写入请求上下文。
//
// Example:
//
//	router.Use(httpmiddleware.Tenant(httpmiddleware.DefaultTenantOptions()))
func Tenant(opts TenantOptions) transportcontract.Middleware {
	opts = normalizeTenantOptions(opts)

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}

			tenant := resolveTenant(c, opts)
			if tenant == "" && opts.Required {
				BadRequest(c, "tenant is required")
				if gc, ok := unwrapGinContext(c); ok {
					gc.Abort()
				}
				return
			}

			if tenant != "" {
				c.Set("tenant", tenant)
				// Also update gin.Request.Context for context.Context value propagation
				if gc, ok := unwrapGinContext(c); ok && gc.Request != nil {
					gc.Request = gc.Request.WithContext(supportcontract.NewTenantContext(gc.Request.Context(), tenant))
				}
				if opts.WriteHeader && opts.HeaderName != "" {
					c.SetHeader(opts.HeaderName, tenant)
				}
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// GetTenant reads the resolved tenant identity from a Gin request context.
//
// GetTenant 从 Gin 请求上下文中读取已解析的租户标识。
func GetTenant(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ""
	}
	if tenant, ok := supportcontract.FromTenantContext(c.Request.Context()); ok {
		return tenant
	}
	return ""
}

// normalizeTenantOptions fills missing tenant options with defaults.
//
// normalizeTenantOptions 用默认值补齐租户选项。
func normalizeTenantOptions(opts TenantOptions) TenantOptions {
	defaults := DefaultTenantOptions()
	if len(opts.ParamKeys) == 0 {
		opts.ParamKeys = defaults.ParamKeys
	}
	if len(opts.HeaderKeys) == 0 {
		opts.HeaderKeys = defaults.HeaderKeys
	}
	if len(opts.QueryKeys) == 0 {
		opts.QueryKeys = defaults.QueryKeys
	}
	if strings.TrimSpace(opts.HeaderName) == "" {
		opts.HeaderName = defaults.HeaderName
	}
	opts.Default = strings.TrimSpace(opts.Default)
	return opts
}

// resolveTenant resolves tenant identity from request param, header, query, or default.
//
// resolveTenant 从请求参数、请求头、查询参数或默认值中解析租户标识。
func resolveTenant(c transportcontract.Context, opts TenantOptions) string {
	for _, key := range opts.ParamKeys {
		if value := strings.TrimSpace(c.Param(key)); value != "" {
			return value
		}
	}
	for _, key := range opts.HeaderKeys {
		if value := strings.TrimSpace(c.GetHeader(key)); value != "" {
			return value
		}
	}
	for _, key := range opts.QueryKeys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			return value
		}
	}
	return opts.Default
}
