package middleware

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// CORSOptions defines the CORS behavior of the HTTP middleware.
//
// CORSOptions 定义 HTTP 中间件的跨域行为。
type CORSOptions struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

// DefaultCORSOptions returns the default CORS configuration.
//
// DefaultCORSOptions 返回默认的跨域配置。
func DefaultCORSOptions() CORSOptions {
	return CORSOptions{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead},
		AllowHeaders:  []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Request-Id", "X-Trace-Id", "X-Idempotency-Key"},
		ExposeHeaders: []string{"Content-Length", "Content-Type", "X-Request-Id", "X-Trace-Id"},
		MaxAgeSeconds: 600,
	}
}

// CORS applies cross-origin resource sharing headers to the HTTP mainline.
//
// CORS 为 HTTP 主线应用跨域响应头。
func CORS(opts CORSOptions) transportcontract.Middleware {
	if len(opts.AllowOrigins) == 0 {
		opts = DefaultCORSOptions()
	}

	// Validate: AllowCredentials + wildcard origin is rejected by browsers.
	// Auto-correct by removing the wildcard and logging a warning.
	// 校验：AllowCredentials + 通配符源 会被浏览器拒绝。
	// 自动修正为移除通配符并记录警告。
	if opts.AllowCredentials {
		hasWildcard := false
		filtered := make([]string, 0, len(opts.AllowOrigins))
		for _, o := range opts.AllowOrigins {
			if strings.TrimSpace(o) == "*" {
				hasWildcard = true
			} else {
				filtered = append(filtered, o)
			}
		}
		if hasWildcard {
			slog.Warn("cors: AllowCredentials=true with AllowOrigins=[\"*\"] is not supported by browsers; removing wildcard origin")
			if len(filtered) > 0 {
				opts.AllowOrigins = filtered
			} else {
				// No valid origins left; disable credentials as fallback.
				// 没有有效的源；禁用 credentials 作为回退。
				opts.AllowCredentials = false
			}
		}
	}

	allowMethods := strings.Join(opts.AllowMethods, ", ")
	allowHeaders := strings.Join(opts.AllowHeaders, ", ")
	exposeHeaders := strings.Join(opts.ExposeHeaders, ", ")
	maxAge := strconv.Itoa(opts.MaxAgeSeconds)

	return func(next transportcontract.Handler) transportcontract.Handler {
		return func(c transportcontract.Context) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}

			origin := c.GetHeader("Origin")
			if origin == "" {
				if next != nil {
					next(c)
				}
				return
			}

			if allowedOrigin, ok := resolveCORSOrigin(origin, opts.AllowOrigins); ok {
				c.SetHeader("Access-Control-Allow-Origin", allowedOrigin)
				c.SetHeader("Vary", "Origin")
			}
			if allowMethods != "" {
				c.SetHeader("Access-Control-Allow-Methods", allowMethods)
			}
			if allowHeaders != "" {
				c.SetHeader("Access-Control-Allow-Headers", allowHeaders)
			}
			if exposeHeaders != "" {
				c.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}
			if opts.AllowCredentials {
				c.SetHeader("Access-Control-Allow-Credentials", "true")
			}
			if opts.MaxAgeSeconds > 0 {
				c.SetHeader("Access-Control-Max-Age", maxAge)
			}

			req := c.Request()
			if req != nil && req.Method == http.MethodOptions {
				c.Status(http.StatusNoContent)
				if gc, ok := unwrapGinContext(c); ok {
					gc.Abort()
				}
				return
			}

			if next != nil {
				next(c)
			}
		}
	}
}

func resolveCORSOrigin(origin string, allowOrigins []string) (string, bool) {
	for _, allowed := range allowOrigins {
		switch strings.TrimSpace(allowed) {
		case "":
			continue
		case "*":
			return "*", true
		default:
			if strings.EqualFold(strings.TrimSpace(allowed), origin) {
				return origin, true
			}
		}
	}
	return "", false
}
