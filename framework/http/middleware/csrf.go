package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// CSRFOptions 定义 CSRF 防护中间件的配置。
//
// 中文说明：
// - TokenLength：生成的 CSRF token 字节数，默认 32；
// - HeaderName：校验 token 的请求头名称，默认 X-CSRF-Token；
// - FormFieldName：表单中 token 字段名，默认 csrf_token；
// - CookieName：存放 token 的 cookie 名称，默认 _csrf；
// - CookiePath：cookie 路径，默认 /；
// - CookieSecure：cookie 是否设置 Secure 标志，默认 false（开发环境 HTTP）；
// - CookieHTTPOnly：cookie 是否设置 HttpOnly 标志，默认 false（前端 JS 需读取）；
// - CookieSameSite：cookie SameSite 属性，默认 Lax；
// - ErrorHandler：自定义错误处理函数，默认返回 403。
type CSRFOptions struct {
	TokenLength    int
	HeaderName     string
	FormFieldName  string
	CookieName     string
	CookiePath     string
	CookieSecure   bool
	CookieHTTPOnly bool
	CookieSameSite http.SameSite
	ErrorHandler   func(transportcontract.HTTPContext)
}

// DefaultCSRFOptions 返回 CSRF 中间件的默认配置。
//
// 中文说明：
// - Token 长度 32 字节，通过请求头 X-CSRF-Token 校验；
// - Cookie 名 _csrf，SameSite=Lax，不设 HttpOnly（前端需读取 token 放入请求头）；
// - 安全方法（GET/HEAD/OPTIONS/TRACE）不校验 token，只设置 cookie。
func DefaultCSRFOptions() CSRFOptions {
	return CSRFOptions{
		TokenLength:    32,
		HeaderName:     "X-CSRF-Token",
		FormFieldName:  "csrf_token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieHTTPOnly: false,
		CookieSameSite: http.SameSiteLaxMode,
	}
}

// CSRF 创建 CSRF 防护中间件。
//
// 中文说明：
// - 对安全方法（GET/HEAD/OPTIONS/TRACE）只设置 cookie，不校验 token；
// - 对非安全方法（POST/PUT/PATCH/DELETE）校验请求头或表单中的 token；
// - token 不匹配时返回 403 Forbidden；
// - 适用场景：传统的 cookie-based 会话 + 表单提交的 Web 应用；
// - 对于纯 API 服务（JWT / Bearer Token 认证），通常不需要 CSRF 防护。
func CSRF(opts CSRFOptions) transportcontract.HTTPMiddleware {
	if opts.TokenLength <= 0 {
		opts.TokenLength = 32
	}
	if opts.HeaderName == "" {
		opts.HeaderName = "X-CSRF-Token"
	}
	if opts.FormFieldName == "" {
		opts.FormFieldName = "csrf_token"
	}
	if opts.CookieName == "" {
		opts.CookieName = "_csrf"
	}
	if opts.CookiePath == "" {
		opts.CookiePath = "/"
	}

	// Warn if CookieSecure is false in production-like environments.
	// CookieSecure=false means the CSRF cookie can be sent over HTTP,
	// which exposes the token to network interception.
	// 警告：CookieSecure=false 时 CSRF cookie 可通过 HTTP 传输，
	// 这会使 token 暴露在网络拦截风险中。
	if !opts.CookieSecure {
		slog.Warn("CSRF middleware configured with CookieSecure=false. " +
			"The CSRF cookie will be sent over HTTP connections, " +
			"exposing the token to network interception. " +
			"Set CookieSecure=true in production environments. " +
			"CSRF 中间件配置了 CookieSecure=false，" +
			"CSRF cookie 将通过 HTTP 连接传输，" +
			"token 可能被网络拦截。生产环境请设置 CookieSecure=true。")
	}

	errorHandler := opts.ErrorHandler
	if errorHandler == nil {
		errorHandler = defaultCSRFErrorHandler
	}

	return func(next transportcontract.HTTPHandler) transportcontract.HTTPHandler {
		return func(c transportcontract.HTTPContext) {
			if c == nil {
				if next != nil {
					next(c)
				}
				return
			}

			req := c.Request()
			if req == nil {
				if next != nil {
					next(c)
				}
				return
			}

			// 从 cookie 中读取已有 token
			cookie, err := req.Cookie(opts.CookieName)
			if err != nil || cookie.Value == "" {
				// 无 token cookie，生成新的并设置
				token := generateCSRFToken(opts.TokenLength)
				setCSRFCookie(c, token, opts)

				// 安全方法不需要校验，直接放行
				if isSafeMethod(req.Method) {
					if next != nil {
						next(c)
					}
					return
				}

				// 非安全方法但没有已有 token，说明是新会话直接操作，拒绝
				errorHandler(c)
				return
			}

			storedToken := cookie.Value

			// 安全方法：只设置 cookie（刷新），不校验
			if isSafeMethod(req.Method) {
				setCSRFCookie(c, storedToken, opts)
				if next != nil {
					next(c)
				}
				return
			}

			// 非安全方法：校验 token
			// 优先从请求头获取，其次从表单字段获取
			submittedToken := c.GetHeader(opts.HeaderName)
			if submittedToken == "" {
				if formValue := req.FormValue(opts.FormFieldName); formValue != "" {
					submittedToken = formValue
				}
			}

			if submittedToken == "" || !secureCompare(submittedToken, storedToken) {
				errorHandler(c)
				return
			}

			if next != nil {
				next(c)
			}
		}
	}
}

// isSafeMethod 判断 HTTP 方法是否为安全方法（不需要 CSRF 校验）。
//
// 中文说明：
// - GET/HEAD/OPTIONS/TRACE 是安全方法，不修改服务器状态；
// - POST/PUT/PATCH/DELETE 是非安全方法，需要校验 CSRF token。
func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

// setCSRFCookie 设置 CSRF token cookie。
//
// 中文说明：
// - 将 token 写入响应 cookie，供后续请求携带；
// - SameSite 默认 Lax，防止跨站请求携带 cookie（但允许顶级导航）；
// - 使用 Add 而非 Set 避免覆盖其他中间件设置的 Set-Cookie 头。
func setCSRFCookie(c transportcontract.HTTPContext, token string, opts CSRFOptions) {
	cookie := &http.Cookie{
		Name:     opts.CookieName,
		Value:    token,
		Path:     opts.CookiePath,
		Secure:   opts.CookieSecure,
		HttpOnly: opts.CookieHTTPOnly,
		SameSite: opts.CookieSameSite,
	}
	// Use Add instead of Set to avoid overwriting other Set-Cookie headers.
	// 使用 Add 而非 Set 避免覆盖其他 Set-Cookie 头。
	if gc, ok := unwrapGinContext(c); ok {
		gc.Writer.Header().Add("Set-Cookie", cookie.String())
	} else {
		c.Header("Set-Cookie", cookie.String())
	}
}

// secureCompare 使用恒定时间比较防止时序攻击。
//
// 中文说明：
// - 使用 crypto/subtle.ConstantTimeCompare 避免通过比较时间推断 token 内容。
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// generateCSRFToken 生成指定长度的随机 CSRF token。
//
// 中文说明：
// - 使用 crypto/rand 生成密码学安全的随机 token；
// - 如果系统熵源不可用，则 panic（这是不可恢复的安全错误）；
// - 编码为 hex 字符串返回。
func generateCSRFToken(length int) string {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand.Read failure means the system entropy source is unavailable.
		// Using all-zero bytes as a CSRF token is a critical security vulnerability,
		// so we must fail loudly rather than silently degrade.
		// crypto/rand.Read 失败意味着系统熵源不可用。
		// 使用全零字节作为 CSRF token 是严重安全漏洞，
		// 必须显式失败而非静默降级。
		panic("csrf: failed to generate secure random token: " + err.Error())
	}
	return hex.EncodeToString(buf)
}

// defaultCSRFErrorHandler 默认的 CSRF 校验失败处理。
//
// 中文说明：
// - 返回 403 Forbidden 状态码；
// - 响应体为 JSON 格式的错误信息。
func defaultCSRFErrorHandler(c transportcontract.HTTPContext) {
	c.JSON(http.StatusForbidden, map[string]string{
		"error": "CSRF token mismatch",
	})
	if gc, ok := unwrapGinContext(c); ok {
		gc.Abort()
	}
}
