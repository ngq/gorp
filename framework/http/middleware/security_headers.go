package middleware

import transportcontract "github.com/ngq/gorp/framework/contract/transport"

// SecurityHeadersOptions defines the default HTTP security response headers.
//
// SecurityHeadersOptions 定义默认 HTTP 安全响应头集合。
type SecurityHeadersOptions struct {
	XFrameOptions       string
	XContentTypeOptions string
	ReferrerPolicy      string
	ContentSecurityPolicy string
	PermissionsPolicy   string
}

// DefaultSecurityHeadersOptions returns the default security header set.
//
// DefaultSecurityHeadersOptions 返回默认安全响应头配置。
func DefaultSecurityHeadersOptions() SecurityHeadersOptions {
	return SecurityHeadersOptions{
		XFrameOptions:       "DENY",
		XContentTypeOptions: "nosniff",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
	}
}

// SecurityHeaders writes a small secure-by-default response header set.
//
// SecurityHeaders 输出一组安全默认响应头。
func SecurityHeaders(opts SecurityHeadersOptions) transportcontract.HTTPMiddleware {
	if opts == (SecurityHeadersOptions{}) {
		opts = DefaultSecurityHeadersOptions()
	}

	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c != nil {
				if opts.XFrameOptions != "" {
					c.Header("X-Frame-Options", opts.XFrameOptions)
				}
				if opts.XContentTypeOptions != "" {
					c.Header("X-Content-Type-Options", opts.XContentTypeOptions)
				}
				if opts.ReferrerPolicy != "" {
					c.Header("Referrer-Policy", opts.ReferrerPolicy)
				}
				if opts.ContentSecurityPolicy != "" {
					c.Header("Content-Security-Policy", opts.ContentSecurityPolicy)
				}
				if opts.PermissionsPolicy != "" {
					c.Header("Permissions-Policy", opts.PermissionsPolicy)
				}
			}
			if next != nil {
				next(c)
			}
		}
	}
}
